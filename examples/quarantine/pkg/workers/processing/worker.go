package processing

import (
	"context"

	"github.com/solo-io/autopilot/pkg/utils"

	v1 "github.com/solo-io/autopilot/examples/quarantine/pkg/apis/quarantines/v1"
)

type Worker struct {
	Kube utils.EzKube
}

func (w *Worker) Sync(ctx context.Context, quarantine *v1.Quarantine, inputs Inputs) (Outputs, v1.QuarantinePhase, error) {
	return Outputs{}, v1.QuarantinePhaseFinished, nil
}
