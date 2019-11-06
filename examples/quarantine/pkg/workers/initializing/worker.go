package initializing

import (
	"context"

	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/examples/quarantine/pkg/apis/quarantines/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
}

func (w *Worker) Sync(ctx context.Context, quarantine *v1.Quarantine) (Outputs, v1.QuarantinePhase, *v1.QuarantineStatusInfo, error) {
	panic("implement me!")
}
