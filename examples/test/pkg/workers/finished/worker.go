package finished

import (
	"context"

	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
}

func (w *Worker) Sync(ctx context.Context, test *v1.Test) (v1.TestPhase, *v1.TestStatusInfo, error) {
	return v1.TestPhaseFinished, &v1.TestStatusInfo{}, nil
}
