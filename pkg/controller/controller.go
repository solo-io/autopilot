package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/pkg/utils"
	"go.uber.org/atomic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

type Handler interface {
	// reconcile an object
	// requeue the object if returning an error, or a non-zero "requeue-after" duration
	Reconcile(object ezkube.Object) (time.Duration, error)
}

type FinalizingEventHandler interface {
	Handler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	FinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	Finalize(object ezkube.Object) error
}

// an EventWatcher is a controller-runtime reconciler that
// uses a cache to retrieve the original event that spawned the
// reconcile request
type EventWatcher interface {
	// the user registers their handlers and starts their watches
	// through the event eventController.
	// they can then register the reconciler with a controller.Controller
	reconcile.Reconciler

	// register a watch with the eventController
	// watches cannot currently be disabled / removed except by
	// terminating the parent controller
	Watch( predicates ...predicate.Predicate) error
}

type eventController struct {
	started          atomic.Bool
	ctx              context.Context
	ctl              controller.Controller
	mgr              manager.Manager
	resource         ezkube.Object
	waitForCacheSync func(stop <-chan struct{}) bool
	logger           logr.Logger
	handler          Handler
}

func NewEventController(ctx context.Context, name string, mgr manager.Manager, resource ezkube.Object, eventHandler Handler) (EventWatcher, error) {
	gvk, err := apiutil.GVKForObject(resource, mgr.GetScheme())
	if err != nil {
		return nil, err
	}

	ec := &eventController{
		ctx:      ctx,
		logger:   log.Log.WithName("event-controller").WithValues("controller", name, "kind", gvk).WithName(name),
		handler:  eventHandler,
		mgr:      mgr,
		resource: resource,
	}

	ctl, err := controller.New(name, mgr, controller.Options{
		Reconciler: ec,
	})

	if err != nil {
		return nil, err
	}

	ec.ctl = ctl
	ec.waitForCacheSync = func(stop <-chan struct{}) bool {
		ec.logger.V(1).Info("waiting for cache sync...")
		return mgr.GetCache().WaitForCacheSync(stop)
	}

	return ec, nil
}

func (ec *eventController) Watch(predicates ...predicate.Predicate) error {

	// send us watch events
	if err := ec.ctl.Watch(&source.Kind{Type: ec.resource}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}

	if synced := ec.waitForCacheSync(ec.ctx.Done()); !synced {
		return errors.Errorf("waiting for cache sync failed")
	}

	return nil
}

func (ec *eventController) finalize(obj ezkube.Object, finalizer FinalizingEventHandler, restClient ezkube.RestClient) (reconcile.Result, error) {
	finalizers := obj.GetFinalizers()
	finalizerName := finalizer.FinalizerName()
	if obj.GetDeletionTimestamp().IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent to
		// registering our finalizer.

		if !utils.ContainsString(finalizers, finalizerName) {
			obj.SetFinalizers(append(
				finalizers,
				finalizerName,
			))
			if err := restClient.Update(context.Background(), obj); err != nil {
				return reconcile.Result{}, err
			}
		}
	}
	// The object is being deleted
	if utils.ContainsString(finalizers, finalizerName) {
		// our finalizer is present, so lets handle any external dependency
		if err := finalizer.Finalize(obj); err != nil {
			// if fail to delete the external dependency here, return with error
			// so that it can be retried
			return reconcile.Result{}, err
		}

		// remove our finalizer from the list and update it.
		obj.SetFinalizers(utils.RemoveString(finalizers, finalizerName))
		if err := restClient.Update(context.Background(), obj); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (ec *eventController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := ec.logger.WithValues("event", request)
	logger.V(2).Info("handling event", "event", request)

	// get the object from our cache
	restClient := ezkube.NewRestClient(ec.mgr)

	obj := ec.resource.DeepCopyObject().(ezkube.Object)
	obj.SetName(request.Name)
	obj.SetNamespace(request.Namespace)
	if err := restClient.Get(ec.ctx, obj); err != nil {
		logger.Error(err, "unable to fetch %T %v", obj, request)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// if the handler is a finalizer, check if we need to finalize
	if finalizer, ok := ec.handler.(FinalizingEventHandler); ok {
		finalizers := obj.GetFinalizers()
		finalizerName := finalizer.FinalizerName()
		if obj.GetDeletionTimestamp().IsZero() {
			// The object is not being deleted, so if it does not have our finalizer,
			// then lets add the finalizer and update the object. This is equivalent to
			// registering our finalizer.

			if !utils.ContainsString(finalizers, finalizerName) {
				obj.SetFinalizers(append(
					finalizers,
					finalizerName,
				))
				if err := restClient.Update(context.Background(), obj); err != nil {
					return reconcile.Result{}, err
				}
			}
		}
		// The object is being deleted
		if utils.ContainsString(finalizers, finalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := finalizer.Finalize(obj); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return reconcile.Result{}, err
			}

			// remove our finalizer from the list and update it.
			obj.SetFinalizers(utils.RemoveString(finalizers, finalizerName))
			if err := restClient.Update(context.Background(), obj); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	requeueAfter, err := ec.handler.Reconcile(obj)
	result := reconcile.Result{RequeueAfter: requeueAfter}
	if err != nil {
		logger.Error(err, "handler error. retrying")
		return result, err
	}
	logger.V(2).Info("handler success.", "retry", requeueAfter > 0)

	return result, nil
}
