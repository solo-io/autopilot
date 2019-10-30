package promoting

import (
	"context"
    "github.com/solo-io/autopilot/pkg/utils"

    v1 "github.com/solo-io/autopilot/examples/promota/pkg/apis/canaries/v1"
)

type Worker struct {
    Kube utils.EzKube
}
func (w *Worker) Sync(ctx context.Context, canary *v1.Canary) (Outputs, v1.CanaryPhase, error) {
    panic("implement me!")
}
