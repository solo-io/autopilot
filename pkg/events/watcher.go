package events

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
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
	// through the event watcher.
	// they can then register the reconciler with a controller.Controller
	reconcile.Reconciler

	// register a watch with the watcher
	// watches cannot currently be disabled / removed except by
	// terminating the parent controller
	Watch(resource runtime.Object, eventHandler EventHandler, predicates ...predicate.Predicate) error
}

type watcher struct {
	events Cache
	ctl    controller.Controller

	lock     sync.RWMutex
	handlers map[schema.GroupVersionKind][]EventHandler
}

func NewWatcher(name string, mgr manager.Manager) (EventWatcher, error) {
	w := &watcher{
		events:   NewCache(),
		handlers: make(map[schema.GroupVersionKind][]EventHandler),
	}

	ctl, err := controller.New(name, mgr, controller.Options{
		Reconciler: w,
	})

	if err != nil {
		return nil, err
	}

	w.ctl = ctl

	return w, nil
}

func (w *watcher) Watch(resource runtime.Object, eventHandler EventHandler, predicates ...predicate.Predicate) error {
	// create a source for the resource type
	src := &source.Kind{Type: resource}

	// send watch events to the Cache
	if err := w.ctl.Watch(src, w.events, predicates...); err != nil {
		return err
	}

	// add the handler to our map
	w.lock.Lock()
	// use gvk as the key
	gvk := resource.GetObjectKind().GroupVersionKind()
	handlers := w.handlers[gvk]
	handlers = append(handlers, eventHandler)
	w.handlers[gvk] = handlers
	w.lock.Unlock()

	return nil
}

func (w *watcher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// event key is stored in the request name
	key := request.Name

	event := w.events.Pop(key)

	switch event.EventType {
	case EventTypeCreate:
		gvk := event.CreateEvent.Object.GetObjectKind().GroupVersionKind()
		handlers, ok := w.handlers[gvk]
		if !ok {
			return reconcile.Result{}, errors.Errorf("no handler registered for gvk %v", gvk)
		}
		for _, h := range handlers {
			err := h.Create(event.CreateEvent.Object)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case EventTypeUpdate:
		gvk := event.UpdateEvent.ObjectNew.GetObjectKind().GroupVersionKind()
		handlers, ok := w.handlers[gvk]
		if !ok {
			return reconcile.Result{}, errors.Errorf("no handler registered for gvk %v", gvk)
		}
		for _, h := range handlers {
			err := h.Update(event.UpdateEvent.ObjectOld, event.UpdateEvent.ObjectNew)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case EventTypeDelete:
		gvk := event.DeleteEvent.Object.GetObjectKind().GroupVersionKind()
		handlers, ok := w.handlers[gvk]
		if !ok {
			return reconcile.Result{}, errors.Errorf("no handler registered for gvk %v", gvk)
		}
		for _, h := range handlers {
			err := h.Delete(event.DeleteEvent.Object)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case EventTypeGeneric:
		gvk := event.GenericEvent.Object.GetObjectKind().GroupVersionKind()
		handlers, ok := w.handlers[gvk]
		if !ok {
			return reconcile.Result{}, errors.Errorf("no handler registered for gvk %v", gvk)
		}
		for _, h := range handlers {
			err := h.Generic(event.GenericEvent.Object)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	default:
		panic("invalid event")
	}

	return reconcile.Result{}, nil
}
