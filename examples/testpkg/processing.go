package testpkg

import (
	"context"
	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
	"github.com/solo-io/autopilot/pkg/metrics"
)

func Process(ctx context.Context, test *v1.Test, metrics metrics.Metrics) (v1.TestPhase, error) {
	result, err := metrics.GetRequestSuccessRate(test.Spec.Target.Name, test.Spec.Target.Namespace, "30s")
	if err != nil {
		return v1.TestPhaseProcessing, err
	}

	if result < test.Spec.Threshold {
		return v1.TestPhaseFailed, nil
	}

	return v1.TestPhaseProcessing, nil
}
