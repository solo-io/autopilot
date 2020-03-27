package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Result = reconcile.Result

type Reconciler interface {
	// reconcile an object
	// requeue the object if returning an error, or a non-zero "requeue-after" duration
	Reconcile(object ezkube.Object) (Result, error)
}

type FinalizingReconciler interface {
	Reconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	FinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	Finalize(object ezkube.Object) error
}

// a Reconcile Loop runs resource reconcilers until the context gets cancelled
type Loop interface {
	RunReconciler(ctx context.Context, reconciler Reconciler, predicates ...predicate.Predicate) error
}

type runner struct {
	name     string
	mgr      manager.Manager
	resource ezkube.Object
}

func NewLoop(name string, mgr manager.Manager, resource ezkube.Object) *runner {
	return &runner{name: name, mgr: mgr, resource: resource}
}

type runnerReconciler struct {
	ctx        context.Context
	mgr        manager.Manager
	resource   ezkube.Object
	logger     logr.Logger
	reconciler Reconciler
}

func (ec *runner) RunReconciler(ctx context.Context, reconciler Reconciler, predicates ...predicate.Predicate) error {
	gvk, err := apiutil.GVKForObject(ec.resource, ec.mgr.GetScheme())
	if err != nil {
		return err
	}
	rec := &runnerReconciler{
		logger:     log.Log.WithName("event-controller").WithValues("kind", gvk).WithName(ec.name),
		mgr:        ec.mgr,
		resource:   ec.resource,
		reconciler: reconciler,
	}

	ctl, err := controller.New(ec.name, ec.mgr, controller.Options{
		Reconciler: rec,
	})
	if err != nil {
		return err
	}

	// send us watch events
	if err := ctl.Watch(&source.Kind{Type: ec.resource}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}

	rec.logger.V(1).Info("waiting for cache sync...")
	if synced := ec.mgr.GetCache().WaitForCacheSync(ctx.Done()); !synced {
		return errors.Errorf("waiting for cache sync failed")
	}

	return nil
}

func (ec *runnerReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := ec.logger.WithValues("event", request)
	logger.V(2).Info("handling event", "event", request)

	// get the object from our cache
	restClient := ezkube.NewRestClient(ec.mgr)

	obj := ec.resource.DeepCopyObject().(ezkube.Object)
	obj.SetName(request.Name)
	obj.SetNamespace(request.Namespace)
	if err := restClient.Get(ec.ctx, obj); err != nil {
		logger.Error(err, fmt.Sprintf("unable to fetch %T", obj), "request", request)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// if the handler is a finalizer, check if we need to finalize
	if finalizer, ok := ec.reconciler.(FinalizingReconciler); ok {
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
		} else {

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
	}

	result, err := ec.reconciler.Reconcile(obj)
	if err != nil {
		logger.Error(err, "handler error. retrying")
		return result, err
	}
	logger.V(2).Info("handler success.", "result", result)

	return result, nil
}
