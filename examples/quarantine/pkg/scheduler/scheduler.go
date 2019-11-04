package scheduler

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/solo-io/autopilot/pkg/metrics"
	"github.com/solo-io/autopilot/pkg/utils"

	v1 "github.com/solo-io/autopilot/examples/quarantine/pkg/apis/quarantines/v1"

	config "github.com/solo-io/autopilot/examples/quarantine/pkg/config"
	finalizer "github.com/solo-io/autopilot/examples/quarantine/pkg/finalizer"
	initializing "github.com/solo-io/autopilot/examples/quarantine/pkg/workers/initializing"
	processing "github.com/solo-io/autopilot/examples/quarantine/pkg/workers/processing"
	aliases "github.com/solo-io/autopilot/pkg/aliases"
)

var log = logf.Log.WithName("scheduler")

func AddToManager(ctx context.Context, mgr manager.Manager, namespace string) error {
	scheduler, err := NewScheduler(ctx, mgr, namespace)
	if err != nil {
		return err
	}
	// Create a new controller
	c, err := controller.New("quarantine-controller", mgr, controller.Options{Reconciler: scheduler})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Quarantine
	log.Info("Registering watch for primary resource Quarantine")
	err = c.Watch(&source.Kind{Type: &v1.Quarantine{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployments and requeue the owner Quarantine
	log.Info("Registering watch for primary resource secondary resource Deployments")
	err = c.Watch(&source.Kind{Type: &aliases.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Quarantine{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Services and requeue the owner Quarantine
	log.Info("Registering watch for primary resource secondary resource Services")
	err = c.Watch(&source.Kind{Type: &aliases.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Quarantine{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Quarantine
	log.Info("Registering watch for primary resource secondary resource Pods")
	err = c.Watch(&source.Kind{Type: &aliases.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.Quarantine{},
	})
	if err != nil {
		return err
	}

	return nil

}

var WorkInterval = config.WorkInterval
var FinalizerName = "quarantine-finalizer"

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

	quarantine := &v1.Quarantine{}
	quarantine.Namespace = request.Namespace
	quarantine.Name = request.Name

	kube := utils.NewEzKube(quarantine, s.mgr)

	if err := kube.Get(s.ctx, quarantine); err != nil {
		// garbage collection and finalizers should handle cleaning up after deletion
		if errors.IsNotFound(err) {
			return result, nil
		}
		return result, err
	}
	// examine DeletionTimestamp to determine if object is under deletion
	if quarantine.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !utils.ContainsString(quarantine.Finalizers, FinalizerName) {
			quarantine.Finalizers = append(quarantine.Finalizers, FinalizerName)
			if err := kube.Ensure(s.ctx, quarantine); err != nil {
				return result, err
			}
		}
	} else {
		// The object is being deleted
		if utils.ContainsString(quarantine.Finalizers, FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := (&finalizer.Finalizer{Kube: kube}).Finalize(s.ctx, quarantine); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return result, err
			}

			// remove our finalizer from the list and update it.
			quarantine.Finalizers = utils.RemoveString(quarantine.Finalizers, FinalizerName)
			if err := kube.Ensure(s.ctx, quarantine); err != nil {
				return result, err
			}
		}

		return result, nil
	}

	switch quarantine.Status.Phase {
	case "", v1.QuarantinePhaseInitializing:
		log.Info("Syncing Quarantine %v in phase Initializing", quarantine.Name)
		outputs, nextPhase, err := (&initializing.Worker{Kube: kube}).Sync(s.ctx, quarantine)
		if err != nil {
			return result, err
		}
		for _, out := range outputs.Deployments {
			if err := kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}
		for _, out := range outputs.Services {
			if err := kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}

		quarantine.Status.Phase = nextPhase
		if err := kube.UpdateStatus(s.ctx, quarantine); err != nil {
			return result, err
		}

		return result, err
	case v1.QuarantinePhaseProcessing:
		log.Info("Syncing Quarantine %v in phase Processing", quarantine.Name)
		inputs, err := s.makeProcessingInputs()
		if err != nil {
			return result, err
		}
		outputs, nextPhase, err := (&processing.Worker{Kube: kube}).Sync(s.ctx, quarantine, inputs)
		if err != nil {
			return result, err
		}
		for _, out := range outputs.Pods {
			if err := kube.Ensure(s.ctx, out); err != nil {
				return result, err
			}
		}

		quarantine.Status.Phase = nextPhase
		if err := kube.UpdateStatus(s.ctx, quarantine); err != nil {
			return result, err
		}

		return result, err
	case v1.QuarantinePhaseFinished:
		log.Info("Syncing Quarantine %v in phase Finished", quarantine.Name)
		// end state, do not requeue
		return reconcile.Result{}, nil
	}
	return result, fmt.Errorf("cannot process Quarantine in unknown phase: %v", quarantine.Status.Phase)
}
func (s *Scheduler) makeProcessingInputs() (processing.Inputs, error) {
	var (
		inputs processing.Inputs
		err    error
	)
	inputs.Metrics = s.Metrics

	return inputs, err
}
