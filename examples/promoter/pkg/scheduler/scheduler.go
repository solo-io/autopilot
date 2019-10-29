package scheduler

// todo: generate this whole file

import (
	"context"
	"fmt"
	v1 "github.com/solo-io/autopilot/examples/promoter/pkg/apis/canaries/v1"
	"github.com/solo-io/autopilot/examples/promoter/pkg/workers/initializing"
	"github.com/solo-io/autopilot/pkg/metrics"
	"github.com/solo-io/autopilot/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// Modify the WorkInterval to change the interval at which workers resync
var WorkInterval = time.Second * 5

func ScheduleWorker(ctx context.Context, kube utils.EzKube, observer metrics.Interface, request reconcile.Request) (reconcile.Result, error) {
	result := reconcile.Result{RequeueAfter: WorkInterval}

	var canary v1.Canary
	canary.Namespace = request.Namespace
	canary.Name = request.Name


	if err := kube.Get(ctx, &canary); err != nil {
		return result, err
	}
	switch canary.Status.Phase {
	case v1.CanaryPhaseInitializing:
		inputs, err := makeInitializingInputs(kube)
		if err != nil {
			return result, err
		}
		outputs, nextPhase, err := (&initializing.Worker{Kube:kube}).Sync(ctx, inputs)
		if err != nil {
			return result, err
		}
		for _, out := range outputs.Services {
			if err := kube.Ensure(ctx, out); err != nil {
				return result, err
			}
		}

		canary.Status.Phase = nextPhase
		if err := kube.UpdateStatus(ctx, canary); err != nil {
			return result, err
		}
		return result, nil
	case v1.CanaryPhaseProgressing:
	case v1.CanaryPhasePromoting:
	case v1.CanaryPhaseSucceeded:
		// end state, do not requeue
		return reconcile.Result{}, nil
	case v1.CanaryPhaseFailed:
		// end state, do not requeue
		return reconcile.Result{}, nil
	}
	return result, fmt.Errorf("cannot process Canary in unknown phase: %v", canary.Status.Phase)
}
