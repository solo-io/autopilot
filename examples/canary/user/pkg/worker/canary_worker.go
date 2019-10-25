package worker

import (
	"context"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/solo-io/autopilot/examples/canary/lib/utils"
	v1 "github.com/solo-io/autopilot/examples/canary/pkg/apis/canaries/v1"
	"github.com/solo-io/autopilot/examples/canary/user/pkg/deployer"
	"github.com/solo-io/autopilot/examples/canary/user/pkg/meshrouter"
	meshmetrics "github.com/solo-io/autopilot/examples/canary/user/pkg/metrics"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sort"
	"strings"
	"time"
)

type CanaryWorker struct {
	ezKube        utils.EzKube
	deployer      *deployer.Deployer
	logger        *zap.SugaredLogger
	eventRecorder record.EventRecorder
	meshRouter    *meshrouter.GlooRouter
	observer      *meshmetrics.GlooObserver
}

var PromAddr = os.Getenv("METRICS_SERVER")

func NewCanaryWorker(mgr manager.Manager) *CanaryWorker {
	ezKube := utils.NewEzKube(&v1.Canary{}, mgr)
	logger := buildLogger()

	namespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		panic(err)
	}

	meshRouter, err := meshrouter.NewGlooRouter("gloo:"+namespace, logger, ezKube)
	if err != nil {
		panic(err)
	}

	if PromAddr == "" {
		PromAddr = "http://prometheus:9090"
	}

	prom, err := meshmetrics.NewPrometheusClient(PromAddr, time.Second*30)
	if err != nil {
		panic(err)
	}

	return &CanaryWorker{
		ezKube:   ezKube,
		deployer: deployer.NewDeployer(ezKube, logger),
		logger:   logger,
		eventRecorder: record.NewBroadcaster().NewRecorder(mgr.GetScheme(), corev1.EventSource{
			Component: "canary-worker",
		}),
		meshRouter: meshRouter,
		observer:   meshmetrics.NewGlooObserver(prom),
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
	if err := w.meshRouter.Reconcile(ctx, canary); err != nil {
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
	primaryWeight, canaryWeight, _, err := w.meshRouter.GetRoutes(ctx, canary)
	if err != nil {
		return err
	}

	// TODO more stats
	// w.recorder.SetWeight(canary, primaryWeight, canaryWeight)

	// check if canary analysis should start (canary revision has changes) or continue
	if ok := w.checkCanaryStatus(ctx, canary); !ok {
		return nil
	}

	// check if canary revision changed during analysis
	if restart := w.hasCanaryRevisionChanged(ctx, canary); restart {
		w.recordEventInfof(canary, "New revision detected! Restarting analysis for %s.%s",
			canary.Spec.TargetRef.Name, canary.Namespace)

		// route all traffic back to primary
		primaryWeight = 100
		canaryWeight = 0
		if err := w.meshRouter.SetRoutes(ctx, canary, primaryWeight, canaryWeight, false); err != nil {
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

	//defer func() {
	//	w.recorder.SetDuration(canary, time.Since(begin))
	//}()
	// check canary deployment status
	// check if analysis should be skipped

	if skip := w.shouldSkipAnalysis(ctx, canary, primaryWeight, canaryWeight); skip {
		return nil
	}

	// route all traffic to primary if analysis has succeeded
	if canary.Status.Phase == v1.CanaryPhasePromoting {
		w.recordEventInfof(canary, "Routing all traffic to primary")
		if err := w.meshRouter.SetRoutes(ctx, canary, 100, 0, false); err != nil {
			return err
		}
		//w.recorder.SetWeight(canary, 100, 0)

		// update status phase
		if err := w.deployer.SetStatusPhase(ctx, canary, v1.CanaryPhaseFinalising); err != nil {
			return err
		}

		return nil
	}

	// scale canary to zero if promotion has finished
	if canary.Status.Phase == v1.CanaryPhaseFinalising {
		if err := w.deployer.Scale(ctx, canary, 0); err != nil {
			return err
		}

		// set status to succeeded
		if err := w.deployer.SetStatusPhase(ctx, canary, v1.CanaryPhaseSucceeded); err != nil {
			return err
		}
		//w.recorder.SetStatus(canary, v1.CanaryPhaseSucceeded)
		//w.runPostRolloutHooks(canary, v1.CanaryPhaseSucceeded)
		w.recordEventInfof(canary, "Promotion completed! Scaling down %s.%s", canary.Spec.TargetRef.Name, canary.Namespace)
		//w.sendNotification(canary, "Canary analysis completed successfully, promotion finished.",
		//	false, false)
		return nil
	}

	// check if the number of failed checks reached the threshold
	if canary.Status.Phase == v1.CanaryPhaseProgressing &&
		(canary.Status.FailedChecks >= canary.Spec.CanaryAnalysis.Threshold) {

		if canary.Status.FailedChecks >= canary.Spec.CanaryAnalysis.Threshold {
			w.recordEventWarningf(canary, "Rolling back %s.%s failed checks threshold reached %v",
				canary.Name, canary.Namespace, canary.Status.FailedChecks)
			//w.sendNotification(canary, fmt.Sprintf("Failed checks threshold reached %v", canary.Status.FailedChecks),
			//	false, true)
		}

		w.recordEventWarningf(canary, "Rolling back %s.%s progress deadline exceeded %v",
			canary.Name, canary.Namespace, err)
		//w.sendNotification(canary, fmt.Sprintf("Progress deadline exceeded %v", err),
		//	false, true)

		// route all traffic back to primary
		primaryWeight = 100
		canaryWeight = 0
		if err := w.meshRouter.SetRoutes(ctx, canary, primaryWeight, canaryWeight, false); err != nil {
			return err
		}

		//w.recorder.SetWeight(canary, primaryWeight, canaryWeight)
		w.recordEventWarningf(canary, "Canary failed! Scaling down %s.%s",
			canary.Name, canary.Namespace)

		// shutdown canary
		if err := w.deployer.Scale(ctx, canary, 0); err != nil {
			return err
		}

		// mark canary as failed
		if err := w.deployer.SyncStatus(ctx, canary, v1.CanaryStatus{Phase: v1.CanaryPhaseFailed, CanaryWeight: 0}); err != nil {
			w.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).Errorf("%v", err)
			return err
		}

		//w.recorder.SetStatus(canary, v1.CanaryPhaseFailed)
		//w.runPostRolloutHooks(canary, v1.CanaryPhaseFailed)
		return nil
	}

	// check if the canary success rate is above the threshold
	// skip check if no traffic is routed or mirrored to canary
	if canaryWeight == 0 && canary.Status.Iterations == 0 {
		w.recordEventInfof(canary, "Starting canary analysis for %s.%s", canary.Spec.TargetRef.Name, canary.Namespace)
	} else {
		if ok := w.analyseCanary(canary); !ok {
			if err := w.deployer.SetStatusFailedChecks(ctx, canary, canary.Status.FailedChecks+1); err != nil {
				return err
			}
			return nil
		}
	}

	// strategy: Blue/Green
	if canary.Spec.CanaryAnalysis.Iterations > 0 {
		// increment iterations
		if canary.Spec.CanaryAnalysis.Iterations > canary.Status.Iterations {
			if err := w.deployer.SetStatusIterations(ctx, canary, canary.Status.Iterations+1); err != nil {
				return err
			}
			w.recordEventInfof(canary, "Advance %s.%s canary iteration %v/%v",
				canary.Name, canary.Namespace, canary.Status.Iterations+1, canary.Spec.CanaryAnalysis.Iterations)
			return nil
		}

		// route all traffic to canary - max iterations reached
		if canary.Spec.CanaryAnalysis.Iterations == canary.Status.Iterations {
			w.recordEventInfof(canary, "Routing all traffic to canary")
			if err := w.meshRouter.SetRoutes(ctx, canary, 0, 100, false); err != nil {
				return err
			}
			//w.recorder.SetWeight(canary, 0, 100)

			// increment iterations
			if err := w.deployer.SetStatusIterations(ctx, canary, canary.Status.Iterations+1); err != nil {
				return err
			}
			return nil
		}

		// promote canary - max iterations reached
		if canary.Spec.CanaryAnalysis.Iterations < canary.Status.Iterations {
			w.recordEventInfof(canary, "Copying %s.%s template spec to %s.%s",
				canary.Spec.TargetRef.Name, canary.Namespace, primaryName, canary.Namespace)
			if err := w.deployer.Promote(ctx, canary); err != nil {
				return err
			}

			// update status phase
			if err := w.deployer.SetStatusPhase(ctx, canary, v1.CanaryPhasePromoting); err != nil {
				return err
			}
			return nil
		}

		return nil
	}

	// set max weight default value to 100%
	maxWeight := 100
	if canary.Spec.CanaryAnalysis.MaxWeight > 0 {
		maxWeight = canary.Spec.CanaryAnalysis.MaxWeight
	}

	// strategy: Canary progressive traffic increase
	if canary.Spec.CanaryAnalysis.StepWeight > 0 {
		// increase traffic weight
		if canaryWeight < maxWeight {

			primaryWeight -= canary.Spec.CanaryAnalysis.StepWeight
			if primaryWeight < 0 {
				primaryWeight = 0
			}
			canaryWeight += canary.Spec.CanaryAnalysis.StepWeight
			if canaryWeight > 100 {
				canaryWeight = 100
			}

			if err := w.meshRouter.SetRoutes(ctx, canary, primaryWeight, canaryWeight, false); err != nil {
				return err
			}

			if err := w.deployer.SetStatusWeight(ctx, canary, canaryWeight); err != nil {
				return err
			}

			//w.recorder.SetWeight(canary, primaryWeight, canaryWeight)
			w.recordEventInfof(canary, "Advance %s.%s canary weight %v", canary.Name, canary.Namespace, canaryWeight)
			return nil
		}

		// promote canary - max weight reached
		if canaryWeight >= maxWeight {

			// update primary spec
			w.recordEventInfof(canary, "Copying %s.%s template spec to %s.%s",
				canary.Spec.TargetRef.Name, canary.Namespace, primaryName, canary.Namespace)
			if err := w.deployer.Promote(ctx, canary); err != nil {
				return err
			}

			// update status phase
			if err := w.deployer.SetStatusPhase(ctx, canary, v1.CanaryPhasePromoting); err != nil {
				return err
			}

			return nil
		}

	}

	return nil
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

func (w *CanaryWorker) hasCanaryRevisionChanged(ctx context.Context, canary *v1.Canary) bool {
	if canary.Status.Phase == v1.CanaryPhaseProgressing {
		if diff, _ := w.deployer.HasDeploymentChanged(ctx, canary); diff {
			return true
		}
	}
	return false
}

func (w *CanaryWorker) shouldSkipAnalysis(ctx context.Context, canary *v1.Canary, primaryWeight int, canaryWeight int) bool {
	// route all traffic to primary
	primaryWeight = 100
	canaryWeight = 0
	if err := w.meshRouter.SetRoutes(ctx, canary, primaryWeight, canaryWeight, false); err != nil {
		w.recordEventWarningf(canary, "%v", err)
		return false
	}
	//w.recorder.SetWeight(canary, primaryWeight, canaryWeight)

	// copy spec and configs from canary to primary
	w.recordEventInfof(canary, "Copying %s.%s template spec to %s-primary.%s",
		canary.Spec.TargetRef.Name, canary.Namespace, canary.Spec.TargetRef.Name, canary.Namespace)
	if err := w.deployer.Promote(ctx, canary); err != nil {
		w.recordEventWarningf(canary, "%v", err)
		return false
	}

	// shutdown canary
	if err := w.deployer.Scale(ctx, canary, 0); err != nil {
		w.recordEventWarningf(canary, "%v", err)
		return false
	}

	// update status phase
	if err := w.deployer.SetStatusPhase(ctx, canary, v1.CanaryPhaseSucceeded); err != nil {
		w.recordEventWarningf(canary, "%v", err)
		return false
	}

	// notify
	//w.recorder.SetStatus(canary, v1.CanaryPhaseSucceeded)
	w.recordEventInfof(canary, "Promotion completed! Canary analysis was skipped for %s.%s",
		canary.Spec.TargetRef.Name, canary.Namespace)
	//w.sendNotification(canary, "Canary analysis was skipped, promotion finished.",
	//	false, false)

	return true
}

func (c *CanaryWorker) analyseCanary(r *v1.Canary) bool {
	// create observer based on the mesh provider
	observer := c.observer

	// run metrics checks
	for _, metric := range r.Spec.CanaryAnalysis.Metrics {
		if metric.Interval == "" {
			metric.Interval = "1m"
		}

		if metric.Name == "request-success-rate" {
			val, err := observer.GetRequestSuccessRate(r.Spec.TargetRef.Name, r.Namespace, metric.Interval)
			if err != nil {
				if strings.Contains(err.Error(), "no values found") {
					c.recordEventWarningf(r, "Halt advancement no values found for metric %s probably %s.%s is not receiving traffic",
						metric.Name, r.Spec.TargetRef.Name, r.Namespace)
				} else {
					c.recordEventErrorf(r, "Metrics server %s query failed: %v", PromAddr, err)
				}
				return false
			}
			if float64(metric.Threshold) > val {
				c.recordEventWarningf(r, "Halt %s.%s advancement success rate %.2f%% < %v%%",
					r.Name, r.Namespace, val, metric.Threshold)
				return false
			}

			//c.recordEventInfof(r, "Check %s passed %.2f%% > %v%%", metric.Name, val, metric.Threshold)
		}

		if metric.Name == "request-duration" {
			val, err := observer.GetRequestDuration(r.Spec.TargetRef.Name, r.Namespace, metric.Interval)
			if err != nil {
				if strings.Contains(err.Error(), "no values found") {
					c.recordEventWarningf(r, "Halt advancement no values found for metric %s probably %s.%s is not receiving traffic",
						metric.Name, r.Spec.TargetRef.Name, r.Namespace)
				} else {
					c.recordEventErrorf(r, "Metrics server %s query failed: %v", PromAddr, err)
				}
				return false
			}
			t := time.Duration(metric.Threshold) * time.Millisecond
			if val > t {
				c.recordEventWarningf(r, "Halt %s.%s advancement request duration %v > %v",
					r.Name, r.Namespace, val, t)
				return false
			}

			//c.recordEventInfof(r, "Check %s passed %v < %v", metric.Name, val, metric.Threshold)
		}

		// custom checks
		if metric.Query != "" {
			val, err := observer.Client.RunQuery(metric.Query)
			if err != nil {
				if strings.Contains(err.Error(), "no values found") {
					c.recordEventWarningf(r, "Halt advancement no values found for metric %s probably %s.%s is not receiving traffic",
						metric.Name, r.Spec.TargetRef.Name, r.Namespace)
				} else {
					c.recordEventErrorf(r, "Metrics server %s query failed for %s: %v", PromAddr, metric.Name, err)
				}
				return false
			}
			if val > float64(metric.Threshold) {
				c.recordEventWarningf(r, "Halt %s.%s advancement %s %.2f > %v",
					r.Name, r.Namespace, metric.Name, val, metric.Threshold)
				return false
			}
		}
	}

	return true
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
