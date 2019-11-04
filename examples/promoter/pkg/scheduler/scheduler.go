package scheduler

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1 "github.com/solo-io/autopilot/examples/promoter/pkg/apis/canaries/v1"

	"github.com/solo-io/autopilot/examples/promoter/pkg/config"
	"github.com/solo-io/autopilot/pkg/aliases"

	"github.com/solo-io/autopilot/examples/promoter/pkg/workers/initializing"
	"github.com/solo-io/autopilot/examples/promoter/pkg/workers/progressing"
	"github.com/solo-io/autopilot/examples/promoter/pkg/workers/promoting"
	"github.com/solo-io/autopilot/pkg/metrics"
)

func AddToManager(ctx context.Context, mgr manager.Manager, namespace string) error {
	scheduler, err := NewScheduler(ctx, mgr, namespace)
	if err != nil {
		return err
	}
	// Create a new controller
	c, err := controller.New("canary-controller", mgr, controller.Options{Reconciler: scheduler})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Canary
	err = c.Watch(&source.Kind{Type: &v1.Canary{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployments and requeue the owner Canary
	err = c.Watch(&source.Kind{Type: &aliases.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Canary{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Services and requeue the owner Canary
	err = c.Watch(&source.Kind{Type: &aliases.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Canary{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource TrafficSplits and requeue the owner Canary
	err = c.Watch(&source.Kind{Type: &aliases.TrafficSplit{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Canary{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource TrafficSplits and requeue the owner Canary
	err = c.Watch(&source.Kind{Type: &aliases.TrafficSplit{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Canary{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployments and requeue the owner Canary
	err = c.Watch(&source.Kind{Type: &aliases.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Canary{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource TrafficSplits and requeue the owner Canary
	err = c.Watch(&source.Kind{Type: &aliases.TrafficSplit{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Canary{},
	})
	if err != nil {
		return err
	}

	return nil

}

var WorkInterval = config.WorkInterval

type Scheduler struct {
	ctx       context.Context
	mgr       manager.Manager
	Metrics   metrics.Metrics
	namespace string
}

func NewScheduler(ctx context.Context, mgr manager.Manager, namespace string) (*Scheduler, error) {
	metricsFactory, err := metrics.NewFactory(config.MetricsServer, config.MeshProvider, time.Second*30)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		ctx:       ctx,
		mgr:       mgr,
		Metrics:   metricsFactory.Observer(),
		namespace: namespace,
	}, nil
}

func (s *Scheduler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	result := reconcile.Result{RequeueAfter: WorkInterval}

	canary := &v1.Canary{}
	canary.Namespace = request.Namespace
	canary.Name = request.Name

	if err := s.kube.Get(s.ctx, canary); err != nil {
		return result, err
	}
	switch canary.Status.Phase {
	case "", v1.CanaryPhaseInitializing:
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
