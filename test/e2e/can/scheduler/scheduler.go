package scheduler

import (
	"context"
	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/controller"
	"github.com/solo-io/autopilot/pkg/request"
	"github.com/solo-io/autopilot/pkg/workqueue"
	"github.com/solo-io/autopilot/test/e2e/can/worker"
	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var logger = log.Log.WithName("canary-scheduler")

type CanaryScheduler struct {
	// The name of the Cluster this Scheduler will watch
	Cluster string

	// root context for the controller
	// used for logging and Kube API requests
	Ctx context.Context

	// Active requests for Canaries across schedulers (clusters)
	ActiveCanaries *request.MultiClusterRequests

	// The Active work queues for all the active schedulers
	ActiveWorkQueues *workqueue.MultiClusterQueues

	// The user-implemented Worker that processes canaries
	Worker worker.CanaryWorker
}

func (s *CanaryScheduler) AddToManager(mgr manager.Manager) error {
	ctl := &controller.Controller{
		Cluster:         s.Cluster,
		Ctx:             s.Ctx,
		PrimaryResource: &v1.CanaryDeployment{},
		InputResources: map[runtime.Object][]predicate.Predicate{
			&v12.Deployment{}: {CanaryPhaseMatches(
				v1.CanaryDeploymentPhaseInitializing,
				v1.CanaryDeploymentPhaseWaiting,
				v1.CanaryDeploymentPhaseEvaluating,
				v1.CanaryDeploymentPhasePromoting,
			)},
		},
		ActivePrimaryResources: s.ActiveCanaries,
		ActiveWorkQueues:       s.ActiveWorkQueues,
		Reconcile:              s.Reconcile,
	}
	return ctl.AddToManager(mgr)
}

func (s *CanaryScheduler) Reconcile(primaryResource runtime.Object) (reconcile.Result, error) {
	canary, ok := primaryResource.(*v1.CanaryDeployment)
	if !ok {
		return reconcile.Result{}, errors.Errorf("Cannot reconcile request for %T, must be *v1.CanaryDeployment", primaryResource)
	}
	switch canary.Status.Phase {
	case v1.CanaryDeploymentPhaseInitializing:
		s.Worker.CanaryInitializing()
	}
}

func CanaryPhaseMatches(phases ...v1.CanaryDeploymentPhase) predicate.Predicate {

	matchesPhase := func(obj runtime.Object) bool {
		canary, ok := obj.(*v1.CanaryDeployment)
		if !ok {
			logger.Error(nil, "cannot process event, expected *v1.CanaryDeployment, got %T", obj)
			return false
		}
		for _, phase := range phases {
			if phase == canary.Status.Phase {
				return true
			}
		}
		return false
	}
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return matchesPhase(event.Object)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return matchesPhase(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return matchesPhase(event.ObjectNew)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return matchesPhase(event.Object)
		},
	}
}
