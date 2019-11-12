package scheduler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes"

	"k8s.io/apimachinery/pkg/api/errors"

	ctl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/solo-io/autopilot/pkg/config"
	"github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/pkg/metrics"
	"github.com/solo-io/autopilot/pkg/scheduler"

	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"

	canarydeploymentmetrics "github.com/solo-io/autopilot/test/e2e/canary/pkg/metrics"
	evaluating "github.com/solo-io/autopilot/test/e2e/canary/pkg/workers/evaluating"
	initializing "github.com/solo-io/autopilot/test/e2e/canary/pkg/workers/initializing"
	promoting "github.com/solo-io/autopilot/test/e2e/canary/pkg/workers/promoting"
	rollback "github.com/solo-io/autopilot/test/e2e/canary/pkg/workers/rollback"
	waiting "github.com/solo-io/autopilot/test/e2e/canary/pkg/workers/waiting"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func AddToManager(params scheduler.Params) error {
	scheduler, err := NewScheduler(params)
	if err != nil {
		return err
	}
	// Create a new controller
	c, err := controller.New("canaryDeployment-controller", params.Manager, controller.Options{Reconciler: scheduler})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CanaryDeployment
	params.Logger.Info("Registering watch for primary resource CanaryDeployment")
	err = c.Watch(&source.Kind{Type: &v1.CanaryDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployments and requeue the owner CanaryDeployment
	params.Logger.Info("Registering watch for secondary resource Deployments")
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.CanaryDeployment{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Services and requeue the owner CanaryDeployment
	params.Logger.Info("Registering watch for secondary resource Services")
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.CanaryDeployment{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource VirtualServices and requeue the owner CanaryDeployment
	params.Logger.Info("Registering watch for secondary resource VirtualServices")
	err = c.Watch(&source.Kind{Type: &istiov1alpha3.VirtualService{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.CanaryDeployment{},
	})
	if err != nil {
		return err
	}

	return nil

}

type Scheduler struct {
	ctx          context.Context
	mgr          manager.Manager
	namespace    string
	logger       logr.Logger
	metrics      canarydeploymentmetrics.CanaryDeploymentMetrics
	workInterval time.Duration
}

func NewScheduler(params scheduler.Params) (*Scheduler, error) {
	cfg := config.ConfigFromContext(params.Ctx)
	metricsServer := metrics.GetMetricsServerAddr(cfg.MeshProvider, cfg.ControlPlaneNs)
	metricsBase, err := metrics.NewPrometheusClient(metricsServer)
	if err != nil {
		return nil, err
	}
	metricsClient := canarydeploymentmetrics.NewMetricsClient(metricsBase)

	workInterval, err := ptypes.Duration(cfg.WorkInterval)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		ctx:          params.Ctx,
		mgr:          params.Manager,
		namespace:    params.Namespace,
		logger:       params.Logger,
		metrics:      metricsClient,
		workInterval: workInterval,
	}, nil
}

func (s *Scheduler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	result := reconcile.Result{RequeueAfter: s.workInterval}

	canaryDeployment := &v1.CanaryDeployment{}
	canaryDeployment.Namespace = request.Namespace
	canaryDeployment.Name = request.Name

	client := ezkube.NewClient(s.mgr)

	if err := client.Get(s.ctx, canaryDeployment); err != nil {
		// garbage collection and finalizers should handle cleaning up after deletion
		if errors.IsNotFound(err) {
			return result, nil
		}
		return result, fmt.Errorf("failed to retrieve requested CanaryDeployment: %v", err)
	}

	// store original status for comparison after sync
	status := canaryDeployment.Status

	logger := s.logger.WithValues(
		"canaryDeployment", canaryDeployment.Namespace+"."+canaryDeployment.Name,
		"phase", canaryDeployment.Status.Phase,
	)

	switch canaryDeployment.Status.Phase {
	case "", v1.CanaryDeploymentPhaseInitializing: // begin worker phase
		logger.Info("Syncing CanaryDeployment in phase Initializing", "name", canaryDeployment.Name)

		worker := &initializing.Worker{
			Client: client,
			Logger: logger,
		}
		inputs, err := s.makeInitializingInputs(client)
		if err != nil {
			return result, fmt.Errorf("failed to make InitializingInputs: %v", err)
		}
		outputs, nextPhase, statusInfo, err := worker.Sync(s.ctx, canaryDeployment, inputs)
		if err != nil {
			return result, fmt.Errorf("failed to run worker for phase Initializing: %v", err)
		}
		for _, out := range outputs.Deployments.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output Deployment<%v.%v> for phase Initializing: %v", out.GetNamespace(), out.GetName(), err)
			}
		}
		for _, out := range outputs.Services.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output Service<%v.%v> for phase Initializing: %v", out.GetNamespace(), out.GetName(), err)
			}
		}
		for _, out := range outputs.VirtualServices.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output VirtualService<%v.%v> for phase Initializing: %v", out.GetNamespace(), out.GetName(), err)
			}
		}

		// update the CanaryDeployment status with the worker's results
		canaryDeployment.Status.Phase = nextPhase
		if statusInfo != nil {
			logger.Info("Updating status of primary resource")
			canaryDeployment.Status.CanaryDeploymentStatusInfo = *statusInfo
		}
	case v1.CanaryDeploymentPhaseWaiting: // begin worker phase
		logger.Info("Syncing CanaryDeployment in phase Waiting", "name", canaryDeployment.Name)

		worker := &waiting.Worker{
			Client: client,
			Logger: logger,
		}
		inputs, err := s.makeWaitingInputs(client)
		if err != nil {
			return result, fmt.Errorf("failed to make WaitingInputs: %v", err)
		}
		outputs, nextPhase, statusInfo, err := worker.Sync(s.ctx, canaryDeployment, inputs)
		if err != nil {
			return result, fmt.Errorf("failed to run worker for phase Waiting: %v", err)
		}
		for _, out := range outputs.Deployments.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output Deployment<%v.%v> for phase Waiting: %v", out.GetNamespace(), out.GetName(), err)
			}
		}

		// update the CanaryDeployment status with the worker's results
		canaryDeployment.Status.Phase = nextPhase
		if statusInfo != nil {
			logger.Info("Updating status of primary resource")
			canaryDeployment.Status.CanaryDeploymentStatusInfo = *statusInfo
		}
	case v1.CanaryDeploymentPhaseEvaluating: // begin worker phase
		logger.Info("Syncing CanaryDeployment in phase Evaluating", "name", canaryDeployment.Name)

		worker := &evaluating.Worker{
			Client: client,
			Logger: logger,
		}
		inputs, err := s.makeEvaluatingInputs(client)
		if err != nil {
			return result, fmt.Errorf("failed to make EvaluatingInputs: %v", err)
		}
		outputs, nextPhase, statusInfo, err := worker.Sync(s.ctx, canaryDeployment, inputs)
		if err != nil {
			return result, fmt.Errorf("failed to run worker for phase Evaluating: %v", err)
		}
		for _, out := range outputs.VirtualServices.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output VirtualService<%v.%v> for phase Evaluating: %v", out.GetNamespace(), out.GetName(), err)
			}
		}

		// update the CanaryDeployment status with the worker's results
		canaryDeployment.Status.Phase = nextPhase
		if statusInfo != nil {
			logger.Info("Updating status of primary resource")
			canaryDeployment.Status.CanaryDeploymentStatusInfo = *statusInfo
		}
	case v1.CanaryDeploymentPhasePromoting: // begin worker phase
		logger.Info("Syncing CanaryDeployment in phase Promoting", "name", canaryDeployment.Name)

		worker := &promoting.Worker{
			Client: client,
			Logger: logger,
		}
		inputs, err := s.makePromotingInputs(client)
		if err != nil {
			return result, fmt.Errorf("failed to make PromotingInputs: %v", err)
		}
		outputs, nextPhase, statusInfo, err := worker.Sync(s.ctx, canaryDeployment, inputs)
		if err != nil {
			return result, fmt.Errorf("failed to run worker for phase Promoting: %v", err)
		}
		for _, out := range outputs.Deployments.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output Deployment<%v.%v> for phase Promoting: %v", out.GetNamespace(), out.GetName(), err)
			}
		}
		for _, out := range outputs.VirtualServices.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output VirtualService<%v.%v> for phase Promoting: %v", out.GetNamespace(), out.GetName(), err)
			}
		}

		// update the CanaryDeployment status with the worker's results
		canaryDeployment.Status.Phase = nextPhase
		if statusInfo != nil {
			logger.Info("Updating status of primary resource")
			canaryDeployment.Status.CanaryDeploymentStatusInfo = *statusInfo
		}
	case v1.CanaryDeploymentPhaseRollBack: // begin worker phase
		logger.Info("Syncing CanaryDeployment in phase RollBack", "name", canaryDeployment.Name)

		worker := &rollback.Worker{
			Client: client,
			Logger: logger,
		}
		inputs, err := s.makeRollBackInputs(client)
		if err != nil {
			return result, fmt.Errorf("failed to make RollBackInputs: %v", err)
		}
		outputs, nextPhase, statusInfo, err := worker.Sync(s.ctx, canaryDeployment, inputs)
		if err != nil {
			return result, fmt.Errorf("failed to run worker for phase RollBack: %v", err)
		}
		for _, out := range outputs.Deployments.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output Deployment<%v.%v> for phase RollBack: %v", out.GetNamespace(), out.GetName(), err)
			}
		}
		for _, out := range outputs.VirtualServices.Items {
			if err := client.Ensure(s.ctx, canaryDeployment, &out); err != nil {
				return result, fmt.Errorf("failed to write output VirtualService<%v.%v> for phase RollBack: %v", out.GetNamespace(), out.GetName(), err)
			}
		}

		// update the CanaryDeployment status with the worker's results
		canaryDeployment.Status.Phase = nextPhase
		if statusInfo != nil {
			logger.Info("Updating status of primary resource")
			canaryDeployment.Status.CanaryDeploymentStatusInfo = *statusInfo
		} // end worker phase

	default:
		return result, fmt.Errorf("cannot process CanaryDeployment in unknown phase: %v", canaryDeployment.Status.Phase)
	}

	canaryDeployment.Status.ObservedGeneration = canaryDeployment.Generation

	if !reflect.DeepEqual(status, canaryDeployment.Status) {
		if err := client.UpdateStatus(s.ctx, canaryDeployment); err != nil {
			return result, fmt.Errorf("failed to update CanaryDeploymentStatus: %v", err)
		}
	}

	return result, nil
}

func (s *Scheduler) makeInitializingInputs(client ezkube.Client) (initializing.Inputs, error) {
	var (
		inputs initializing.Inputs
		err    error
	)
	err = client.List(s.ctx, &inputs.Deployments, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}

	return inputs, err
}

func (s *Scheduler) makeWaitingInputs(client ezkube.Client) (waiting.Inputs, error) {
	var (
		inputs waiting.Inputs
		err    error
	)
	err = client.List(s.ctx, &inputs.Deployments, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}

	return inputs, err
}

func (s *Scheduler) makeEvaluatingInputs(client ezkube.Client) (evaluating.Inputs, error) {
	var (
		inputs evaluating.Inputs
		err    error
	)
	inputs.Metrics = s.metrics
	err = client.List(s.ctx, &inputs.VirtualServices, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}

	return inputs, err
}

func (s *Scheduler) makePromotingInputs(client ezkube.Client) (promoting.Inputs, error) {
	var (
		inputs promoting.Inputs
		err    error
	)
	err = client.List(s.ctx, &inputs.Deployments, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}
	err = client.List(s.ctx, &inputs.VirtualServices, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}

	return inputs, err
}

func (s *Scheduler) makeRollBackInputs(client ezkube.Client) (rollback.Inputs, error) {
	var (
		inputs rollback.Inputs
		err    error
	)
	err = client.List(s.ctx, &inputs.Deployments, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}
	err = client.List(s.ctx, &inputs.VirtualServices, ctl.InNamespace(s.namespace))
	if err != nil {
		return inputs, err
	}

	return inputs, err
}