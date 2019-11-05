package finalizer

import (
	"context"

	"github.com/solo-io/autopilot/pkg/utils"

	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
)

type Finalizer struct {
	Kube utils.EzKube
}

func (f *Finalizer) Finalize(ctx context.Context, test *v1.Test) error {
	return nil
}
