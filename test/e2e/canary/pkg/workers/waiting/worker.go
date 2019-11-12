package waiting

import (
	"context"
	"github.com/solo-io/autopilot/test/e2e/canary/pkg/weights"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/test/e2e/canary/pkg/parameters"

	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
	Logger logr.Logger
}

func (w *Worker) Sync(ctx context.Context, canary *v1.CanaryDeployment, inputs Inputs) (Outputs, v1.CanaryDeploymentPhase, *v1.CanaryDeploymentStatusInfo, error) {
	canaryName := canary.Name + "-canary"
	logger := w.Logger.WithValues("canaryName", canaryName)

	logger.Info("checking canary for changes")

	canaryDeployment, ok := inputs.FindDeployment(canaryName, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("deployment %v not found for canary %v", canaryName, canary.Name)
	}

	targetDeployment, ok := inputs.FindDeployment(canary.Name, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("primary deployment not found for canary %v", canary.Name)
	}

	if deploymentsEqual(targetDeployment.Spec, canaryDeployment.Spec) {
		w.Logger.Info("canary has not changed")
		return Outputs{}, v1.CanaryDeploymentPhaseWaiting, nil, nil
	}

	logger.Info("diff", "canary", canaryDeployment.Spec.Template, "target", targetDeployment.Spec.Template)

	logger.Info("updating traffic split, updating and scaling up canary", "replicas", 1)

	targetDeployment.Spec.Template.Labels = canaryDeployment.Spec.Template.Labels
	targetDeployment.Spec.Selector = canaryDeployment.Spec.Selector

	canaryDeployment.Spec = targetDeployment.Spec
	canaryDeployment.Spec.Replicas = pointer.Int32Ptr(1) // scale up canary

	virtualService, ok := inputs.FindVirtualService(canary.Name, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("virtual service not found for canary %v", canary.Name)
	}

	// kick off the analysis with 10% split
	if err := weights.StepWeights(&virtualService, 10); err != nil {
		return Outputs{}, "", nil, errors.Wrapf(err, "failed to step virtual service weights for canary %v", canary.Name)
	}

	return Outputs{
		Deployments: parameters.Deployments{
			Items: []appsv1.Deployment{
				canaryDeployment,
			},
		},
		VirtualServices: parameters.VirtualServices{
			Items: []v1alpha3.VirtualService{
				virtualService,
			},
		},
	},
		v1.CanaryDeploymentPhaseEvaluating,
		&v1.CanaryDeploymentStatusInfo{TimeStarted: metav1.Now(), History: canary.Status.History},
		nil
}

func deploymentsEqual(target, canary appsv1.DeploymentSpec) bool {
	// ignore labels as these are overridden by our controller
	target.Selector = canary.Selector
	target.Template.Labels = canary.Template.Labels

	return reflect.DeepEqual(target, canary)
}
