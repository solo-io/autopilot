package finalizer

import (
	"context"

	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/examples/quarantine/pkg/apis/quarantines/v1"
)

type Finalizer struct {
	Client ezkube.Client
}

func (f *Finalizer) Finalize(ctx context.Context, quarantine *v1.Quarantine) error {
	panic("implement me!")
}
