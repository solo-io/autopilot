package processing

import (
	"context"
	"time"

	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
}

func (w *Worker) Sync(ctx context.Context, test *v1.Test, inputs Inputs) (v1.TestPhase, *v1.TestStatusInfo, error) {
	successRate, err := inputs.Metrics.GetRequestSuccessRate(test.Spec.Target.Name, test.Spec.Target.Namespace, "1m")
	if err != nil {
		return "", nil, err
	}

	if successRate < test.Spec.Threshold {
		return v1.TestPhaseFailed, nil, nil
	}

	if time.Now().Sub(test.Status.TimeStarted.Time) >= test.Spec.Timeout.Duration {
		return v1.TestPhaseFinished, nil, nil
	}

	return v1.TestPhaseProcessing, nil, nil
}
