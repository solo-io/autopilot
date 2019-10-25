package worker

import (
	"context"
	"fmt"
	"github.com/solo-io/autopilot/examples/canary/lib/utils"
	v1 "github.com/solo-io/autopilot/examples/canary/pkg/apis/canaries/v1"
	"github.com/solo-io/autopilot/examples/canary/user/pkg/deployer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sort"
)

type CanaryWorker struct {
	ezKube        utils.EzKube
	deployer      *deployer.Deployer
	logger        *zap.SugaredLogger
	eventRecorder record.EventRecorder
}

func NewCanaryWorker(mgr manager.Manager) *CanaryWorker {
	ezKube := utils.NewEzKube(&v1.Canary{}, mgr)
	logger := buildLogger()

	return &CanaryWorker{
		ezKube:   ezKube,
		deployer: deployer.NewDeployer(ezKube, logger),
		logger:   logger,
		eventRecorder: record.NewBroadcaster().NewRecorder(mgr.GetScheme(), corev1.EventSource{
			Component: "canary-worker",
		}),
	}
}

func buildLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevel()
	log, _ := config.Build()
	return log.Sugar()
}

// this method must be named DoWork and accept Context and Canary as its only argument
func (w *CanaryWorker) DoWork(ctx context.Context, canary *v1.Canary) {
	if err := w.advanceCanary(ctx, canary); err != nil {
		w.recordEventErrorf(canary, "%v", err)
	}
}

func (w *CanaryWorker) advanceCanary(ctx context.Context, canary *v1.Canary) error {
	// set status condition for new canaries
	if canary.Status.Conditions == nil {
		if ok, conditions := w.deployer.MakeStatusConditions(canary.Status, v1.CanaryPhaseInitializing); ok {
			cdCopy := canary.DeepCopy()
			cdCopy.Status.Conditions = conditions
			cdCopy.Status.LastTransitionTime = metav1.Now()
			cdCopy.Status.Phase = v1.CanaryPhaseInitializing
			return w.ezKube.UpdateStatus(ctx, cdCopy)
		}
	}

	// create primary deployment
	labelSelector, ports, err := w.deployer.Initialize(ctx, canary)
	if err != nil {
		return err
	}

	targetName := canary.Spec.TargetRef.Name
	primaryName := fmt.Sprintf("%s-primary", targetName)
	canaryName := fmt.Sprintf("%s-canary", targetName)

	// canary svc
	err = w.reconcileService(ctx, canary, canaryName, targetName, labelSelector, ports)
	if err != nil {
		return err
	}

	// primary svc
	err = w.reconcileService(ctx, canary, primaryName, primaryName, labelSelector, ports)
	if err != nil {
		return err
	}

	// create or update virtual service
	if err := meshRouter.Reconcile(ctx, canary); err != nil {
		return err
	}

	shouldAdvance, err := w.shouldAdvance(ctx, canary)
	if err != nil {
		return err
	}

	if !shouldAdvance {
		// TODO: metrics w.recorder.SetStatus(canary, canary.Status.Phase)
		return nil
	}

	// get the routing settings
	primaryWeight, canaryWeight, mirrored, err := meshRouter.GetRoutes(ctx, canary)
	if err != nil {
		return err
	}

	// TODO more stats
	// c.recorder.SetWeight(canary, primaryWeight, canaryWeight)

	// check if canary analysis should start (canary revision has changes) or continue
	if ok := w.checkCanaryStatus(ctx, canary); !ok {
		return nil
	}

	// set max weight default value to 100%
	maxWeight := 100
	if canary.Spec.CanaryAnalysis.MaxWeight > 0 {
		maxWeight = canary.Spec.CanaryAnalysis.MaxWeight
	}

	// check if canary revision changed during analysis
	if restart := w.hasCanaryRevisionChanged(ctx, canary); restart {
		w.recordEventInfof(canary, "New revision detected! Restarting analysis for %s.%s",
			canary.Spec.TargetRef.Name, canary.Namespace)

		// route all traffic back to primary
		primaryWeight = 100
		canaryWeight = 0
		if err := meshRouter.SetRoutes(ctx, canary, primaryWeight, canaryWeight, false); err != nil {
			return err
		}

		// reset status
		status := v1.CanaryStatus{
			Phase:        v1.CanaryPhaseProgressing,
			CanaryWeight: 0,
			FailedChecks: 0,
			Iterations:   0,
		}
		if err := w.deployer.SyncStatus(ctx, canary, status); err != nil {
			return err
		}

		return nil
	}

}

func (w *CanaryWorker) shouldAdvance(ctx context.Context, canary *v1.Canary) (bool, error) {
	if canary.Status.LastAppliedSpec == "" ||
		canary.Status.Phase == v1.CanaryPhaseInitializing ||
		canary.Status.Phase == v1.CanaryPhaseProgressing ||
		canary.Status.Phase == v1.CanaryPhaseWaiting ||
		canary.Status.Phase == v1.CanaryPhasePromoting ||
		canary.Status.Phase == v1.CanaryPhaseFinalising {
		return true, nil
	}

	newDep, err := w.deployer.HasDeploymentChanged(ctx, canary)
	if err != nil {
		return false, err
	}
	return newDep, nil
}

func (w *CanaryWorker) reconcileService(ctx context.Context,
	canary *v1.Canary, name, target, labelSelector string, ports map[string]int32) error {
	portName := canary.Spec.Service.PortName
	if portName == "" {
		portName = "http"
	}

	targetPort := intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: canary.Spec.Service.Port,
	}

	if canary.Spec.Service.TargetPort.String() != "0" {
		targetPort = canary.Spec.Service.TargetPort
	}

	svcSpec := corev1.ServiceSpec{
		Type:     corev1.ServiceTypeClusterIP,
		Selector: map[string]string{labelSelector: target},
		Ports: []corev1.ServicePort{
			{
				Name:       portName,
				Protocol:   corev1.ProtocolTCP,
				Port:       canary.Spec.Service.Port,
				TargetPort: targetPort,
			},
		},
	}

	for n, p := range ports {
		cp := corev1.ServicePort{
			Name:     n,
			Protocol: corev1.ProtocolTCP,
			Port:     p,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: p,
			},
		}

		svcSpec.Ports = append(svcSpec.Ports, cp)
	}

	sort.SliceStable(svcSpec.Ports, func(i, j int) bool {
		return svcSpec.Ports[i].Port < svcSpec.Ports[j].Port
	})

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: canary.Namespace,
			Labels:    map[string]string{labelSelector: name},
		},
		Spec: svcSpec,
	}

	return w.ezKube.Ensure(ctx, svc)
}

func (w *CanaryWorker) checkCanaryStatus(ctx context.Context, canary *v1.Canary) bool {
	// TODO metrics w.recorder.SetStatus(canary, canary.Status.Phase)
	if canary.Status.Phase == v1.CanaryPhaseProgressing ||
		canary.Status.Phase == v1.CanaryPhasePromoting ||
		canary.Status.Phase == v1.CanaryPhaseFinalising {
		return true
	}

	if canary.Status.Phase == "" || canary.Status.Phase == v1.CanaryPhaseInitializing {
		if err := w.deployer.SyncStatus(ctx, canary, v1.CanaryStatus{Phase: v1.CanaryPhaseInitialized}); err != nil {
			w.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).Errorf("%v", err)
			return false
		}
		//w.recorder.SetStatus(canary, spec.CanaryPhaseInitialized)
		//w.recordEventInfof(canary, "Initialization done! %s.%s", canary.Name, canary.Namespace)
		//w.sendNotification(canary, "New deployment detected, initialization completed.",
		//	true, false)
		return false
	}

	w.recordEventInfof(canary, "New revision detected! Scaling up %s.%s", canary.Spec.TargetRef.Name, canary.Namespace)
	//w.sendNotification(canary, "New revision detected, starting canary analysis.",
	//	true, false)
	if err := w.deployer.Scale(ctx, canary, 1); err != nil {
		w.recordEventErrorf(canary, "%v", err)
		return false
	}
	if err := w.deployer.SyncStatus(ctx, canary, v1.CanaryStatus{Phase: v1.CanaryPhaseProgressing}); err != nil {
		w.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).Errorf("%v", err)
		return false
	}
	//w.recorder.SetStatus(canary, spec.CanaryPhaseProgressing)
	return false
}

func (w *CanaryWorker) hasCanaryRevisionChanged(ctx context.Context, cd *v1.Canary) bool {
	if cd.Status.Phase == v1.CanaryPhaseProgressing {
		if diff, _ := w.deployer.HasDeploymentChanged(ctx, cd); diff {
			return true
		}
	}
	return false
}

// graceful teardown if necessary
// objects created with controller references will be garbage collected and do not need to be removed here
func (w *CanaryWorker) Teardown(ctx context.Context, canary *v1.Canary) {}

func (w *CanaryWorker) recordEventInfof(r *v1.Canary, template string, args ...interface{}) {
	w.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).Infof(template, args...)
	w.eventRecorder.Event(r, corev1.EventTypeNormal, "Synced", fmt.Sprintf(template, args...))
}

func (w *CanaryWorker) recordEventErrorf(r *v1.Canary, template string, args ...interface{}) {
	w.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).Errorf(template, args...)
	w.eventRecorder.Event(r, corev1.EventTypeWarning, "Synced", fmt.Sprintf(template, args...))
}

func (w *CanaryWorker) recordEventWarningf(r *v1.Canary, template string, args ...interface{}) {
	w.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).Infof(template, args...)
	w.eventRecorder.Event(r, corev1.EventTypeWarning, "Synced", fmt.Sprintf(template, args...))
}
