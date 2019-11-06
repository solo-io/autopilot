package finalizer

import (
	"context"

	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Finalizer struct {
	Client ezkube.Client
}

func (f *Finalizer) Finalize(ctx context.Context, test *v1.Test) error {
	return nil
}
