package initializing

import (
	"context"
	v1 "github.com/solo-io/autopilot/examples/promoter/pkg/apis/canaries/v1"
	"github.com/solo-io/autopilot/pkg/utils"
)

type Worker struct {
	kube utils.EzKube
}

func (w *Worker) Sync(ctx context.Context, inputs Inputs) (Outputs, v1.CanaryPhase, error) {
	return Outputs{}, "", nil
}
