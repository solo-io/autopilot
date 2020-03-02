package events

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type EventHandler interface {
	Create(object runtime.Object) error

	Delete(object runtime.Object) error

	Update(old, new runtime.Object) error

	Generic(object runtime.Object) error
}

// an EventWatcher is a controller-runtime reconciler that
// uses a cache to retrieve the original event that spawned the
// reconcile request
type EventWatcher interface {
	// the user registers their handlers and starts their watches
	// through the event EventWatcher.
	// they can then register the reconciler with a controller.Controller
	reconcile.Reconciler

	// start a watch with the EventWatcher
	// watches cannot currently be disabled / removed except by
	// terminating the parent controller
	Watch(ctx context.Context, eventHandler EventHandler, predicates ...predicate.Predicate) error
}

type watcher struct {
	name         string          // name of this watch/controller
	mgr          manager.Manager // manager
	resource     runtime.Object  // resource type
}

func NewWatcher(name string, mgr manager.Manager, resource runtime.Object) *watcher {
	return &watcher{name: name, mgr: mgr, resource: resource}
}

func (w *watcher) Watch(ctx context.Context, eventHandler EventHandler, predicates ...predicate.Predicate) error {
	reconciler := &eventWatcher{
		events:       NewCache(),
		eventHandler: eventHandler,
	}

	ctl, err := controller.New(w.name, w.mgr, controller.Options{
		Reconciler: reconciler,
	})
	if err != nil {
		return err
	}

	// create a source for the resource type
	src := &source.Kind{Type: w.resource}

	// send watch events to the Cache
	if err := ctl.Watch(src, reconciler.events, predicates...); err != nil {
		return err
	}

	if synced := w.mgr.GetCache().WaitForCacheSync(ctx.Done()); !synced {
		return errors.Errorf("waiting for cache sync failed")
	}

	return nil
}

type eventWatcher struct {
	events       Cache
	eventHandler EventHandler
}

func (w *eventWatcher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// event key is stored in the request name
	key := request.Name
	log.Log.V(4).Info("event eventWatcher reconciling event", "key", key)

	event := w.events.Get(key)

	if event == nil {
		return reconcile.Result{}, errors.Errorf("internal error: received invalid event key %v", key)
	}

	switch evt := event.(type) {
	case createEvent:
		if err := w.eventHandler.Create(evt.Object); err != nil {
			return reconcile.Result{}, err
		}
	case updateEvent:
		if err := w.eventHandler.Update(evt.ObjectOld, evt.ObjectNew); err != nil {
			return reconcile.Result{}, err
		}
	case deleteEvent:
		if err := w.eventHandler.Delete(evt.Object); err != nil {
			return reconcile.Result{}, err
		}
	case genericEvent:
		if err := w.eventHandler.Generic(evt.Object); err != nil {
			return reconcile.Result{}, err
		}
	default:
		panic("invalid event")
	}

	w.events.Forget(key)

	return reconcile.Result{}, nil
}
