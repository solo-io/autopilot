package waiting

import (
	"context"
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

	if reflect.DeepEqual(canaryDeployment.Spec.Template, targetDeployment.Spec.Template) {
		w.Logger.Info("canary has not changed")
		return Outputs{}, v1.CanaryDeploymentPhaseWaiting, nil, nil
	}

	logger.Info("diff", "canary", canaryDeployment.Spec.Template, "target", targetDeployment.Spec.Template)

	logger.Info("updating canary deployment and scaling up", "replicas", 1)

	canaryDeployment.Spec = targetDeployment.Spec
	canaryDeployment.Spec.Replicas = pointer.Int32Ptr(1) // scale up canary

	return Outputs{
		Deployments: parameters.Deployments{
			Items: []appsv1.Deployment{
				canaryDeployment,
			},
		},
	},
		v1.CanaryDeploymentPhaseEvaluating,
		&v1.CanaryDeploymentStatusInfo{TimeStarted: metav1.Now()},
		nil
}
