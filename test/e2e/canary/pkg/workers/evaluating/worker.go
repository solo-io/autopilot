package evaluating

import (
	"context"
	"github.com/solo-io/autopilot/test/e2e/canary/pkg/weights"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/test/e2e/canary/pkg/parameters"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"

	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
	Logger logr.Logger
}

// amount to shift weight from primary to canary for each sync
var StepWeight = int32(5)

func (w *Worker) Sync(ctx context.Context, canary *v1.CanaryDeployment, inputs Inputs) (Outputs, v1.CanaryDeploymentPhase, *v1.CanaryDeploymentStatusInfo, error) {
	w.Logger.Info("evaluating canary metrics", "successThreshold", canary.Spec.SuccessThreshold)

	canaryName := canary.Name + "-canary"

	val, err := inputs.Metrics.GetIstioSuccessRate(ctx, canary.Namespace, canaryName, "1m")
	if err != nil {
		return Outputs{}, "", nil, errors.Errorf("failed to get metrics for canary deployment %v", canaryName)
	}

	successRateScalar, ok := val.Value.(*model.Scalar)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("wrong metrics value, type %T (expected *model.Scalar): %v", val.Value, val.Value)
	}

	successRate := float64(successRateScalar.Value)
	w.Logger.Info("observed success rate", "successRate", successRate)

	if successRate < canary.Spec.SuccessThreshold {
		w.Logger.Info("success rate below threshold, rolling back...")
		return Outputs{}, v1.CanaryDeploymentPhaseRollBack, nil, nil
	}

	timeLeft := canary.Spec.AnalysisPeriod.Duration - time.Now().Sub(canary.Status.TimeStarted.Time)

	if timeLeft <= 0 {
		w.Logger.Info("success rate maintained above threshold for analysis period, promoting...")
		return Outputs{}, v1.CanaryDeploymentPhasePromoting, nil, nil
	}

	w.Logger.Info("continuing analysis period... shifting %v% more traffic to canary...", "timeLeft", timeLeft)
	virtualService, ok := inputs.FindVirtualService(canary.Name, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("virtual service not found for canary %v", canary.Name)
	}

	primaryWeight, canaryWeight, err := weights.GetWeights(virtualService)
	if err != nil {
		return Outputs{}, "", nil, errors.Wrapf(err, "failed to get virtual service weights for canary %v", canary.Name)
	}

	// once primary weight is below 0, we just monitor all the traffic going to canary
	if primaryWeight > 0 {
		if err := weights.SetWeights(&virtualService, primaryWeight-StepWeight, canaryWeight+StepWeight); err != nil {
			return Outputs{}, "", nil, errors.Wrapf(err, "failed to set weights for virtual service for canary %v", canary.Name)
		}
	}

	// we still want to be in evaluating phase while we are processing
	return Outputs{VirtualServices: parameters.VirtualServices{
		Items: []v1alpha3.VirtualService{virtualService},
	}},
		v1.CanaryDeploymentPhaseEvaluating, nil, nil
}

// increments the weight to the canary destination, decrements the weight to the primary destination
func stepWeights(virtualService *v1alpha3.VirtualService) error {
	// for shift weights on each per-port route
	for _, httpPort := range virtualService.Spec.Http {
		if len(httpPort.Route) != 2 {
			return errors.Errorf("expected 2 routes on http rule %v, found %v", httpPort.Name, len(virtualService.Spec.Http))
		}
		primaryRoute, canaryRoute := httpPort.Route[0], httpPort.Route[1]

		// decrement primary
		primaryRoute.Weight -= StepWeight
		if primaryRoute.Weight < 0 {
			primaryRoute.Weight = 0
		}

		// increment canary
		canaryRoute.Weight += StepWeight
		if primaryRoute.Weight > 100 {
			primaryRoute.Weight = 100
		}
	}

	return nil
}
