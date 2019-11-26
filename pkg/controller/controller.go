package controller

import (
	"context"
	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/ezkube"
	aphandler "github.com/solo-io/autopilot/pkg/handler"
	"github.com/solo-io/autopilot/pkg/request"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// the Controller is an opinionated wrapper for the controller runtime Controller
// it handles creating watches for a top level resource, its dependencies inputs and outputs
type Controller struct {
	// name of the controller (used for logging)
	Name string

	// root context for the controller
	// used for logging and Kube API requests
	Ctx context.Context

	// the reconcile function to call. it will always call with a request for a top-level resource
	Reconcile func(topLevelResource runtime.Object) (reconcile.Result, error)

	// the type of top-level resource to watch
	TopLevelResource ezkube.Object

	// optional predicates for filtering on the top-level resource
	TopLevelPredicates []predicate.Predicate

	// map of input resources to optional predicates for each one
	// Reconcile will be called for all controlling objects for any input resource
	InputResources map[runtime.Object][]predicate.Predicate

	// map of output resources to optional predicates for each one
	// Reconcile will be called for the controlling object for any output resource
	OutputResources map[runtime.Object][]predicate.Predicate
}

// Add the Controller to
func (o *Controller) AddToManager(mgr manager.Manager) error {
	client := ezkube.NewRestClient(mgr)

	reconcileFunc := reconcile.Func(func(request reconcile.Request) (reconcile.Result, error) {
		topLevelResource, ok := o.TopLevelResource.DeepCopyObject().(ezkube.Object)
		if !ok {
			return reconcile.Result{}, errors.Errorf("unexpected object type %T, did not contain metadata", o.TopLevelResource)
		}
		topLevelResource.SetName(request.Name)
		topLevelResource.SetNamespace(request.Namespace)
		if err := client.Get(o.Ctx, topLevelResource); err != nil {
			return reconcile.Result{}, err
		}
		return o.Reconcile(topLevelResource)
	})

	ctl, err := controller.New(o.Name, mgr, controller.Options{Reconciler: reconcileFunc})
	if err != nil {
		return err
	}

	// create a container for sharing requests for owner resource
	// the owner resource watch will populate these requests
	// the input resource watches will fire these requests
	activeOwnerRequests := &request.SyncRequests{}

	// add the predicate that tracks each live request (for existing instances of the top level resource)
	predicates := append(o.TopLevelPredicates, &aphandler.RequestTrackingPredicate{Requests: activeOwnerRequests})

	// start the top-level watch
	if err := ctl.Watch(&source.Kind{Type: o.TopLevelResource}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}

	// set up watches for input resources.
	// will enqueue requests for all top-level resources
	for input, predicates := range o.InputResources {
		if err := ctl.Watch(
			&source.Kind{Type: input},
			&aphandler.EnqueueStaticRequests{RequestsToEnqueue: activeOwnerRequests},
			predicates...,
		); err != nil {
			return err
		}
	}

	// set up watches for output resources with controller refs
	for output, predicates := range o.OutputResources {
		if err := ctl.Watch(
			&source.Kind{Type: output},
			&handler.EnqueueRequestForOwner{OwnerType: o.TopLevelResource, IsController: true},
			predicates...,
		); err != nil {
			return err
		}
	}

	return nil
}
