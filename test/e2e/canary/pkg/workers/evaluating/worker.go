package evaluating

import (
	"context"
	"strings"
	"time"

	"github.com/solo-io/autopilot/test/e2e/canary/pkg/weights"

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

	// format interval string to avoid error
	interval := strings.TrimSuffix(canary.Spec.MeasurementInterval.Duration.String(), "0s")
	interval = strings.TrimSuffix(interval, "0m")

	val, err := inputs.Metrics.GetIstioSuccessRate(ctx, canary.Namespace, canaryName, interval)
	if err != nil {
		return Outputs{}, "", nil, errors.Errorf("failed to get metrics for canary deployment %v", canaryName)
	}

	var successRate float64
	switch val := val.Value.(type) {
	case *model.Scalar:
		successRate = float64(val.Value)
	case model.Vector:
		if len(val) < 1 {
			return Outputs{}, "", nil, errors.Errorf("invalid metrics value, expected at least 1 result: %v", val)
		}
		successRate = float64(val[0].Value)
	default:
		return Outputs{}, "", nil, errors.Errorf("wrong metrics value, type %T (expected *model.Scalar or model.Vector): %v", val, val)
	}

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

	if err := weights.StepWeights(&virtualService, StepWeight); err != nil {
		return Outputs{}, "", nil, errors.Wrapf(err, "failed to step virtual service weights for canary %v", canary.Name)
	}

	// we still want to be in evaluating phase while we are processing
	return Outputs{VirtualServices: parameters.VirtualServices{
			Items: []v1alpha3.VirtualService{virtualService},
		}},
		v1.CanaryDeploymentPhaseEvaluating, nil, nil
}
