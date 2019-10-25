package deployer

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/mitchellh/hashstructure"
	ex "github.com/pkg/errors"
	"github.com/solo-io/autopilot/examples/canary/lib/utils"
	v1 "github.com/solo-io/autopilot/examples/canary/pkg/apis/canaries/v1"
	"go.uber.org/zap"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
)

// Deployer is managing the operations for Kubernetes deployment kind
type Deployer struct {
	EzKube utils.EzKube
	Logger *zap.SugaredLogger
	Labels []string
}

func NewDeployer(ezKube utils.EzKube, logger *zap.SugaredLogger) *Deployer {
	return &Deployer{EzKube: ezKube, Logger: logger, Labels: []string{"app"}}
}

// Initialize creates the primary deployment, hpa,
// scales to zero the canary deployment and returns the pod selector label and container ports
func (c *Deployer) Initialize(ctx context.Context, cd *v1.Canary) (label string, ports map[string]int32, err error) {
	primaryName := fmt.Sprintf("%s-primary", cd.Spec.TargetRef.Name)
	label, ports, err = c.createPrimaryDeployment(ctx, cd)
	if err != nil {
		return "", ports, fmt.Errorf("creating deployment %s.%s failed: %v", primaryName, cd.Namespace, err)
	}

	if cd.Status.Phase == "" || cd.Status.Phase == v1.CanaryPhaseInitializing {
		c.Logger.With("canary", fmt.Sprintf("%s.%s", cd.Name, cd.Namespace)).Infof("Scaling down %s.%s", cd.Spec.TargetRef.Name, cd.Namespace)
		if err := c.Scale(ctx, cd, 0); err != nil {
			return "", ports, err
		}
	}

	return label, ports, nil
}

// Promote copies the pod spec, secrets and config maps from canary to primary
func (c *Deployer) Promote(ctx context.Context, cd *v1.Canary) error {
	targetName := cd.Spec.TargetRef.Name
	primaryName := fmt.Sprintf("%s-primary", targetName)

	canary, err := c.EzKube.GetDeployment(ctx, cd.Namespace, targetName)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("deployment %s.%s not found", targetName, cd.Namespace)
		}
		return fmt.Errorf("deployment %s.%s query error %v", targetName, cd.Namespace, err)
	}

	label, err := c.getSelectorLabel(canary)
	if err != nil {
		return fmt.Errorf("invalid label selector! Deployment %s.%s spec.selector.matchLabels must contain selector 'app: %s'",
			targetName, cd.Namespace, targetName)
	}

	primary, err := c.EzKube.GetDeployment(ctx, cd.Namespace, primaryName)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("deployment %s.%s not found", primaryName, cd.Namespace)
		}
		return fmt.Errorf("deployment %s.%s query error %v", primaryName, cd.Namespace, err)
	}

	primaryCopy := primary.DeepCopy()
	primaryCopy.Spec.ProgressDeadlineSeconds = canary.Spec.ProgressDeadlineSeconds
	primaryCopy.Spec.MinReadySeconds = canary.Spec.MinReadySeconds
	primaryCopy.Spec.RevisionHistoryLimit = canary.Spec.RevisionHistoryLimit
	primaryCopy.Spec.Strategy = canary.Spec.Strategy

	// update spec with primary secrets and config maps
	primaryCopy.Spec.Template.Spec = canary.Spec.Template.Spec

	// update pod annotations to ensure a rolling update
	annotations, err := c.makeAnnotations(canary.Spec.Template.Annotations)
	if err != nil {
		return err
	}
	primaryCopy.Spec.Template.Annotations = annotations

	primaryCopy.Spec.Template.Labels = makePrimaryLabels(canary.Spec.Template.Labels, primaryName, label)

	// apply update
	err = c.EzKube.Ensure(ctx, primaryCopy)
	if err != nil {
		return fmt.Errorf("updating deployment %s.%s template spec failed: %v",
			primaryCopy.GetName(), primaryCopy.Namespace, err)
	}

	return nil
}

// HasDeploymentChanged returns true if the canary deployment pod spec has changed
func (c *Deployer) HasDeploymentChanged(ctx context.Context, cd *v1.Canary) (bool, error) {
	targetName := cd.Spec.TargetRef.Name
	canary, err := c.EzKube.GetDeployment(ctx, cd.Namespace, targetName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, fmt.Errorf("deployment %s.%s not found", targetName, cd.Namespace)
		}
		return false, fmt.Errorf("deployment %s.%s query error %v", targetName, cd.Namespace, err)
	}

	if cd.Status.LastAppliedSpec == "" {
		return true, nil
	}

	newHash, err := hashstructure.Hash(canary.Spec.Template, nil)
	if err != nil {
		return false, fmt.Errorf("hash error %v", err)
	}

	// do not trigger a canary deployment on manual rollback
	if cd.Status.LastPromotedSpec == fmt.Sprintf("%d", newHash) {
		return false, nil
	}

	if cd.Status.LastAppliedSpec != fmt.Sprintf("%d", newHash) {
		return true, nil
	}

	return false, nil
}

// Scale sets the canary deployment replicas
func (c *Deployer) Scale(ctx context.Context, cd *v1.Canary, replicas int32) error {
	targetName := cd.Spec.TargetRef.Name
	dep, err := c.EzKube.GetDeployment(ctx, cd.Namespace, targetName)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("deployment %s.%s not found", targetName, cd.Namespace)
		}
		return fmt.Errorf("deployment %s.%s query error %v", targetName, cd.Namespace, err)
	}

	depCopy := dep.DeepCopy()
	depCopy.Spec.Replicas = int32p(replicas)

	err = c.EzKube.Ensure(ctx, depCopy)
	if err != nil {
		return fmt.Errorf("scaling %s.%s to %v failed: %v", depCopy.GetName(), depCopy.Namespace, replicas, err)
	}
	return nil
}

func (c *Deployer) createPrimaryDeployment(ctx context.Context, cd *v1.Canary) (string, map[string]int32, error) {
	targetName := cd.Spec.TargetRef.Name
	primaryName := fmt.Sprintf("%s-primary", cd.Spec.TargetRef.Name)

	canaryDep, err := c.EzKube.GetDeployment(ctx, cd.Namespace, targetName)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", nil, fmt.Errorf("deployment %s.%s not found, retrying", targetName, cd.Namespace)
		}
		return "", nil, err
	}

	label, err := c.getSelectorLabel(canaryDep)
	if err != nil {
		return "", nil, fmt.Errorf("invalid label selector! Deployment %s.%s spec.selector.matchLabels must contain selector 'app: %s'",
			targetName, cd.Namespace, targetName)
	}

	var ports map[string]int32
	if cd.Spec.Service.PortDiscovery {
		p, err := c.getPorts(cd, canaryDep)
		if err != nil {
			return "", nil, fmt.Errorf("port discovery failed with error: %v", err)
		}
		ports = p
	}

	primaryDep, err := c.EzKube.GetDeployment(ctx, cd.Namespace, primaryName)
	if errors.IsNotFound(err) {
		annotations, err := c.makeAnnotations(canaryDep.Spec.Template.Annotations)
		if err != nil {
			return "", nil, err
		}

		replicas := int32(1)
		if canaryDep.Spec.Replicas != nil && *canaryDep.Spec.Replicas > 0 {
			replicas = *canaryDep.Spec.Replicas
		}

		// create primary deployment
		primaryDep = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      primaryName,
				Namespace: cd.Namespace,
				Labels: map[string]string{
					label: primaryName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				ProgressDeadlineSeconds: canaryDep.Spec.ProgressDeadlineSeconds,
				MinReadySeconds:         canaryDep.Spec.MinReadySeconds,
				RevisionHistoryLimit:    canaryDep.Spec.RevisionHistoryLimit,
				Replicas:                int32p(replicas),
				Strategy:                canaryDep.Spec.Strategy,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						label: primaryName,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      makePrimaryLabels(canaryDep.Spec.Template.Labels, primaryName, label),
						Annotations: annotations,
					},
					// update spec with the primary secrets and config maps
					Spec: canaryDep.Spec.Template.Spec,
				},
			},
		}

		if err := c.EzKube.Ensure(ctx, primaryDep); err != nil {
			return "", nil, err
		}

		c.Logger.With("canary", fmt.Sprintf("%s.%s", cd.Name, cd.Namespace)).Infof("Deployment %s.%s created", primaryDep.GetName(), cd.Namespace)
	}

	return label, ports, nil
}

// SyncStatus encodes the canary pod spec and updates the canary status
func (c *Deployer) SyncStatus(ctx context.Context, cd *v1.Canary, status v1.CanaryStatus) error {
	dep, err := c.EzKube.GetDeployment(ctx, cd.Namespace, cd.Spec.TargetRef.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("deployment %s.%s not found", cd.Spec.TargetRef.Name, cd.Namespace)
		}
		return ex.Wrap(err, "SyncStatus deployment query error")
	}

	hash, err := hashstructure.Hash(dep.Spec.Template, nil)
	if err != nil {
		return ex.Wrap(err, "SyncStatus hash error")
	}

	firstTry := true
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		var selErr error
		if !firstTry {
			selErr = c.EzKube.Get(ctx, cd)
			if selErr != nil {
				return selErr
			}
		}
		cdCopy := cd.DeepCopy()
		cdCopy.Status.Phase = status.Phase
		cdCopy.Status.CanaryWeight = status.CanaryWeight
		cdCopy.Status.FailedChecks = status.FailedChecks
		cdCopy.Status.Iterations = status.Iterations
		cdCopy.Status.LastAppliedSpec = fmt.Sprintf("%d", hash)
		cdCopy.Status.LastTransitionTime = metav1.Now()

		if ok, conditions := c.MakeStatusConditions(cd.Status, status.Phase); ok {
			cdCopy.Status.Conditions = conditions
		}

		err = c.EzKube.UpdateStatus(ctx, cdCopy)
		firstTry = false
		return
	})
	if err != nil {
		return ex.Wrap(err, "SyncStatus")
	}
	return nil
}

// MakeStatusCondition updates the canary status conditions based on canary phase
func (c *Deployer) MakeStatusConditions(canaryStatus v1.CanaryStatus,
	phase v1.CanaryPhase) (bool, []v1.CanaryCondition) {
	currentCondition := c.getStatusCondition(canaryStatus, v1.PromotedType)

	message := "New deployment detected, starting initialization."
	status := corev1.ConditionUnknown
	switch phase {
	case v1.CanaryPhaseInitializing:
		status = corev1.ConditionUnknown
		message = "New deployment detected, starting initialization."
	case v1.CanaryPhaseInitialized:
		status = corev1.ConditionTrue
		message = "Deployment initialization completed."
	case v1.CanaryPhaseWaiting:
		status = corev1.ConditionUnknown
		message = "Waiting for approval."
	case v1.CanaryPhaseProgressing:
		status = corev1.ConditionUnknown
		message = "New revision detected, starting canary analysis."
	case v1.CanaryPhasePromoting:
		status = corev1.ConditionUnknown
		message = "Canary analysis completed, starting primary rolling update."
	case v1.CanaryPhaseFinalising:
		status = corev1.ConditionUnknown
		message = "Canary analysis completed, routing all traffic to primary."
	case v1.CanaryPhaseSucceeded:
		status = corev1.ConditionTrue
		message = "Canary analysis completed successfully, promotion finished."
	case v1.CanaryPhaseFailed:
		status = corev1.ConditionFalse
		message = "Canary analysis failed, deployment scaled to zero."
	}

	newCondition := &v1.CanaryCondition{
		Type:               v1.PromotedType,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Message:            message,
		Reason:             string(phase),
	}

	if currentCondition != nil &&
		currentCondition.Status == newCondition.Status &&
		currentCondition.Reason == newCondition.Reason {
		return false, nil
	}

	if currentCondition != nil && currentCondition.Status == newCondition.Status {
		newCondition.LastTransitionTime = currentCondition.LastTransitionTime
	}

	return true, []v1.CanaryCondition{*newCondition}
}

// SetStatusPhase updates the canary status phase
func (c *Deployer) SetStatusPhase(ctx context.Context, cd *v1.Canary, phase v1.CanaryPhase) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		var selErr error
		if !firstTry {
			selErr = c.EzKube.Get(ctx, cd)
			if selErr != nil {
				return selErr
			}
		}
		cdCopy := cd.DeepCopy()
		cdCopy.Status.Phase = phase
		cdCopy.Status.LastTransitionTime = metav1.Now()

		if phase != v1.CanaryPhaseProgressing && phase != v1.CanaryPhaseWaiting {
			cdCopy.Status.CanaryWeight = 0
			cdCopy.Status.Iterations = 0
		}

		// on promotion set primary spec hash
		if phase == v1.CanaryPhaseInitialized || phase == v1.CanaryPhaseSucceeded {
			cdCopy.Status.LastPromotedSpec = cd.Status.LastAppliedSpec
		}

		if ok, conditions := c.MakeStatusConditions(cdCopy.Status, phase); ok {
			cdCopy.Status.Conditions = conditions
		}

		err = c.EzKube.UpdateStatus(ctx, cdCopy)
		firstTry = false
		return
	})
	if err != nil {
		return ex.Wrap(err, "SetStatusPhase")
	}
	return nil
}

// SetStatusFailedChecks updates the canary failed checks counter
func (c *Deployer) SetStatusFailedChecks(ctx context.Context, cd *v1.Canary, val int) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		var selErr error
		if !firstTry {
			selErr = c.EzKube.Get(ctx, cd)
			if selErr != nil {
				return selErr
			}
		}
		cdCopy := cd.DeepCopy()
		cdCopy.Status.FailedChecks = val
		cdCopy.Status.LastTransitionTime = metav1.Now()

		err = c.EzKube.UpdateStatus(ctx, cdCopy)
		firstTry = false
		return
	})
	if err != nil {
		return ex.Wrap(err, "SetStatusFailedChecks")
	}
	return nil
}

// SetStatusWeight updates the canary status weight value
func (c *Deployer) SetStatusWeight(ctx context.Context, cd *v1.Canary, val int) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		var selErr error
		if !firstTry {
			selErr = c.EzKube.Get(ctx, cd)
			if selErr != nil {
				return selErr
			}
		}
		cdCopy := cd.DeepCopy()
		cdCopy.Status.CanaryWeight = val
		cdCopy.Status.LastTransitionTime = metav1.Now()

		err = c.EzKube.UpdateStatus(ctx, cdCopy)
		firstTry = false
		return
	})
	if err != nil {
		return ex.Wrap(err, "SetStatusWeight")
	}
	return nil
}

// SetStatusIterations updates the canary status iterations value
func (c *Deployer) SetStatusIterations(ctx context.Context, cd *v1.Canary, val int) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		var selErr error
		if !firstTry {
			selErr = c.EzKube.Get(ctx, cd)
			if selErr != nil {
				return selErr
			}
		}

		cdCopy := cd.DeepCopy()
		cdCopy.Status.Iterations = val
		cdCopy.Status.LastTransitionTime = metav1.Now()

		err = c.EzKube.UpdateStatus(ctx, cdCopy)
		firstTry = false
		return
	})

	if err != nil {
		return ex.Wrap(err, "SetStatusIterations")
	}
	return nil
}

// GetStatusCondition returns a condition based on type
func (c *Deployer) getStatusCondition(status v1.CanaryStatus, conditionType v1.CanaryConditionType) *v1.CanaryCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

// makeAnnotations appends an unique ID to annotations map
func (c *Deployer) makeAnnotations(annotations map[string]string) (map[string]string, error) {
	idKey := "canary-id"
	res := make(map[string]string)
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return res, err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	id := fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	for k, v := range annotations {
		if k != idKey {
			res[k] = v
		}
	}
	res[idKey] = id

	return res, nil
}

// getSelectorLabel returns the selector match label
func (c *Deployer) getSelectorLabel(deployment *appsv1.Deployment) (string, error) {
	for _, l := range c.Labels {
		if _, ok := deployment.Spec.Selector.MatchLabels[l]; ok {
			return l, nil
		}
	}

	return "", fmt.Errorf("selector not found")
}

var sidecars = map[string]bool{
	"istio-proxy": true,
	"envoy":       true,
}

// getPorts returns a list of all container ports
func (c *Deployer) getPorts(cd *v1.Canary, deployment *appsv1.Deployment) (map[string]int32, error) {
	ports := make(map[string]int32)

	for _, container := range deployment.Spec.Template.Spec.Containers {
		// exclude service mesh proxies based on container name
		if _, ok := sidecars[container.Name]; ok {
			continue
		}
		for i, p := range container.Ports {
			// exclude canary.service.port or canary.service.targetPort
			if cd.Spec.Service.TargetPort.String() == "0" {
				if p.ContainerPort == cd.Spec.Service.Port {
					continue
				}
			} else {
				if cd.Spec.Service.TargetPort.Type == intstr.Int {
					if p.ContainerPort == cd.Spec.Service.TargetPort.IntVal {
						continue
					}
				}
				if cd.Spec.Service.TargetPort.Type == intstr.String {
					if p.Name == cd.Spec.Service.TargetPort.StrVal {
						continue
					}
				}
			}
			name := fmt.Sprintf("tcp-%s-%v", container.Name, i)
			if p.Name != "" {
				name = p.Name
			}

			ports[name] = p.ContainerPort
		}
	}

	return ports, nil
}

func makePrimaryLabels(labels map[string]string, primaryName string, label string) map[string]string {
	res := make(map[string]string)
	for k, v := range labels {
		if k != label {
			res[k] = v
		}
	}
	res[label] = primaryName

	return res
}

func int32p(i int32) *int32 {
	return &i
}

func int32Default(i *int32) int32 {
	if i == nil {
		return 1
	}

	return *i
}
