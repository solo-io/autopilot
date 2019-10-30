package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/autopilot/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/solo-io/autopilot/examples/promota/pkg/apis/canaries/v1"

	"github.com/solo-io/autopilot/examples/promota/pkg/workers/initializing"
	"github.com/solo-io/autopilot/examples/promota/pkg/workers/progressing"
	"github.com/solo-io/autopilot/examples/promota/pkg/workers/promoting"
	"github.com/solo-io/autopilot/pkg/metrics"
)

// Modify the WorkInterval to change the interval at which workers resync
var WorkInterval = time.Second * 5

type Scheduler struct {
	ctx       context.Context
	kube      utils.EzKube
	Metrics   metrics.Metrics
	namespace string
}

func NewScheduler(ctx context.Context, kube utils.EzKube, m metrics.Metrics, namespace string) *Scheduler {
	return &Scheduler{
		ctx:       ctx,
		kube:      kube,
		Metrics:   m,
		namespace: namespace,
	}
}

func (s *Scheduler) ScheduleWorker(request reconcile.Request) (reconcile.Result, error) {
	result := reconcile.Result{RequeueAfter: WorkInterval}

	canary := &v1.Canary{}
	canary.Namespace = request.Namespace
	canary.Name = request.Name

	if err := s.kube.Get(s.ctx, canary); err != nil {
		return result, err
	}
	switch canary.Status.Phase {
	case v1.CanaryPhaseInitializing:
		inputs, err := s.makeInitializingInputs()
		if err != nil {
			return result, err
		}
		outputs, nextPhase, err := (&initializing.Worker{Kube: s.kube}).Sync(s.ctx, canary, inputs)
		if err != nil {
			return result, err
		}
		for _, out := range outputs.Deployments {
			if err := s.kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}
		for _, out := range outputs.Services {
			if err := s.kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}
		for _, out := range outputs.TrafficSplits {
			if err := s.kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}

		canary.Status.Phase = nextPhase
		if err := s.kube.UpdateStatus(s.ctx, canary); err != nil {
			return result, err
		}

		return result, err
	case v1.CanaryPhaseProgressing:
		inputs, err := s.makeProgressingInputs()
		if err != nil {
			return result, err
		}
		outputs, nextPhase, err := (&progressing.Worker{Kube: s.kube}).Sync(s.ctx, canary, inputs)
		if err != nil {
			return result, err
		}
		for _, out := range outputs.TrafficSplits {
			if err := s.kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}

		canary.Status.Phase = nextPhase
		if err := s.kube.UpdateStatus(s.ctx, canary); err != nil {
			return result, err
		}

		return result, err
	case v1.CanaryPhasePromoting:
		outputs, nextPhase, err := (&promoting.Worker{Kube: s.kube}).Sync(s.ctx, canary)
		if err != nil {
			return result, err
		}
		for _, out := range outputs.Deployments {
			if err := s.kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}
		for _, out := range outputs.TrafficSplits {
			if err := s.kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}

		canary.Status.Phase = nextPhase
		if err := s.kube.UpdateStatus(s.ctx, canary); err != nil {
			return result, err
		}

		return result, err
	case v1.CanaryPhaseSucceeded:
		// end state, do not requeue
		return reconcile.Result{}, nil
	case v1.CanaryPhaseFailed:
		// end state, do not requeue
		return reconcile.Result{}, nil
	}
	return result, fmt.Errorf("cannot process Canary in unknown phase: %v", canary.Status.Phase)
}
func (s *Scheduler) makeInitializingInputs() (initializing.Inputs, error) {
	var (
		inputs initializing.Inputs
		err    error
	)
	inputs.Deployments, err = s.kube.ListDeployments(s.ctx, s.namespace)
	if err != nil {
		return inputs, err
	}

	return inputs, err
}
func (s *Scheduler) makeProgressingInputs() (progressing.Inputs, error) {
	var (
		inputs progressing.Inputs
		err    error
	)
	inputs.Metrics = s.Metrics

	return inputs, err
}
