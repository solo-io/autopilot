package controller

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/ezkube"
	aphandler "github.com/solo-io/autopilot/pkg/handler"
	"github.com/solo-io/autopilot/pkg/request"
	"github.com/solo-io/autopilot/pkg/workqueue"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// the Controller is an opinionated wrapper for the controller runtime Controller
// it handles creating watches for a top level resource, its dependencies inputs and outputs
// each Controller runs for a single Primary resource in a single Cluster
// The controller watches inputs and outputs in remote clusters
type Controller struct {
	// The name of the Cluster this controller will watch
	Cluster string

	// root context for the controller
	// used for logging and Kube API requests
	Ctx context.Context

	// the reconcile function to call. it will always call with a request for a top-level resource
	Reconcile func(primaryResource runtime.Object) (reconcile.Result, error)

	// the type of top-level resource to watch
	PrimaryResource ezkube.Object

	// optional predicates for filtering on the top-level resource
	PrimaryPredicates []predicate.Predicate

	// map of input resources to optional predicates for each one
	// Reconcile will be called for all primary resources on an event for an input resource event
	InputResources map[runtime.Object][]predicate.Predicate

	// map of output resources to optional predicates for each one
	// Reconcile will be called for the controlling primary resource for
	// an output resource event
	OutputResources map[runtime.Object][]predicate.Predicate

	// A shared set of requests for active (existing) Primary resources.
	// The controller will track requests for the primary resource
	// in the target cluster.
	// The controller will fire requests for events on input and output resources
	// to all tracked primary resources, across clusters
	ActivePrimaryResources *request.MultiClusterRequests

	// A shared set of queues for all the active clusters
	ActiveWorkQueues *workqueue.MultiClusterQueues
}

// Add the Primary watches for the controller Controller to the Manager
// this adds the watch for the top-level resource as well as
// watches for secondary resources.
// Shared Queues and Requests will be propagated for this cluster/kind
// back to the ActivePrimaryResources and ActiveWorkQueues
func (c *Controller) AddToManager(mgr manager.Manager) error {
	client := ezkube.NewRestClient(mgr)

	schemas, _, err := mgr.GetScheme().ObjectKinds(c.PrimaryResource)
	if err != nil {
		return errors.Wrapf(err, "no schema found for %T", c.PrimaryResource)
	}

	if len(schemas) == 0 {
		return errors.Errorf("empty schema list found for %T", c.PrimaryResource)
	}

	name := fmt.Sprintf("%v,Cluster=%v", schemas[0].String(), c.Cluster)

	logger := log.Log.WithName(name)

	// instantiate the base controller.
	// this will track our metrics
	ctl, err := controller.New(name, mgr, controller.Options{
		Reconciler: c.reconcileFunc(client),
	})
	if err != nil {
		return err
	}

	// start the top-level watch
	logger.Info(fmt.Sprintf("starting primary watch for %T", c.PrimaryPredicates))
	if err := ctl.Watch(&source.Kind{Type: c.PrimaryResource},
		&aphandler.MultiHandler{
			Handlers: []handler.EventHandler{
				// register the queue for the controller to the queue registry
				aphandler.QueueRegisteringHandler(c.Cluster, c.ActiveWorkQueues),

				// track each live request (for existing instances of the top level resource) across clusters
				&aphandler.MultiClusterRequestTracker{
					Cluster:  c.Cluster,
					Requests: c.ActivePrimaryResources,
				},

				// handle the request itself
				&handler.EnqueueRequestForObject{},
			},
		}, c.PrimaryPredicates...); err != nil {
		return err
	}

	// set up watches for input resources.
	// will enqueue requests for all top-level resources
	for input, predicates := range c.InputResources {
		logger.Info(fmt.Sprintf("starting secondary watch for %T", input))
		if err := ctl.Watch(
			&source.Kind{Type: input},
			&aphandler.BroadcastRequests{
				WorkQueues:        c.ActiveWorkQueues,
				RequestsToEnqueue: c.ActivePrimaryResources,
			},
			predicates...,
		); err != nil {
			return err
		}
	}

	// set up watches for output resources with controller refs
	for output, predicates := range c.OutputResources {
		logger.Info(fmt.Sprintf("starting secondary watch for %T", output))
		if err := ctl.Watch(
			&source.Kind{Type: output},
			&handler.EnqueueRequestForOwner{
				OwnerType:    c.PrimaryResource,
				IsController: true,
			},
			predicates...,
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) reconcileFunc(client ezkube.RestClient) reconcile.Func {
	return func(request reconcile.Request) (reconcile.Result, error) {
		primaryResource, ok := c.PrimaryResource.DeepCopyObject().(ezkube.Object)
		if !ok {
			return reconcile.Result{}, errors.Errorf("unexpected object type %T, did not contain metadata", c.PrimaryResource)
		}
		primaryResource.SetName(request.Name)
		primaryResource.SetNamespace(request.Namespace)
		if err := client.Get(c.Ctx, primaryResource); err != nil {
			return reconcile.Result{}, err
		}
		return c.Reconcile(primaryResource)
	}
}
