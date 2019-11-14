package rollback

import (
	"context"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/test/e2e/canary/pkg/parameters"
	"github.com/solo-io/autopilot/test/e2e/canary/pkg/weights"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"

	"github.com/go-logr/logr"
	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
	Logger logr.Logger
}

func (w *Worker) Sync(ctx context.Context, canary *v1.CanaryDeployment, inputs Inputs) (Outputs, v1.CanaryDeploymentPhase, *v1.CanaryDeploymentStatusInfo, error) {
	canaryName := canary.Name + "-canary"

	w.Logger.Info("rolling back canary... scaling down canary deployment and shifting all traffic back to primary...")

	virtualService, ok := inputs.FindVirtualService(canary.Name, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("virtual service not found for canary %v", canary.Name)
	}

	canaryDeployment, ok := inputs.FindDeployment(canaryName, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("canary deployment not found for canary %v", canary.Name)
	}
	canaryDeployment.Spec.Replicas = pointer.Int32Ptr(0)

	if err := weights.SetWeights(&virtualService, 100, 0); err != nil {
		return Outputs{}, "", nil, errors.Wrapf(err, "failed to set weights for virtual service for canary %v", canary.Name)
	}

	// append to the canary's history
	status := canary.Status.CanaryDeploymentStatusInfo
	status.History = append(status.History, v1.CanaryResult{
		PromotionSucceeded: false,
		ObservedGeneration: canaryDeployment.Status.ObservedGeneration,
	})

	// update the deployments, resume waiting
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
	}, v1.CanaryDeploymentPhaseWaiting, &status, nil
}
