package scheduler

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Parameters for initializing the (generated) Autopilot Scheduler
// Users should not need to interact with this package
// unless they wish to manually modify their scheduler
type Params struct {
	// root context
	Ctx context.Context

	// parent manager
	Manager manager.Manager

	// watch namespace for the scheduler
	Namespace string

	// root logger
	Logger logr.Logger
}
