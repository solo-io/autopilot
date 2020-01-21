package events

import (
	"sync"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logr = log.Log.WithName("watcher")

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
	scheme *runtime.Scheme

	lock     sync.RWMutex
	handlers map[schema.GroupVersionKind][]EventHandler
}

func NewWatcher(name string, mgr manager.Manager) (EventWatcher, error) {
	w := &watcher{
		events:   NewCache(),
		handlers: make(map[schema.GroupVersionKind][]EventHandler),
		scheme:   mgr.GetScheme(),
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

func (w *watcher) getGvk(resource runtime.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := w.scheme.ObjectKinds(resource)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	if len(gvks) < 1 {
		return schema.GroupVersionKind{}, errors.Errorf("no gvk registered for resource %T", resource)
	}
	if len(gvks) > 1 {
		logr.V(4).Info("multiple versions registered for type, defaulting to first",
			"type", resource, "gvks", gvks)
	}

	return gvks[0], nil
}

func (w *watcher) getHandlers(resource runtime.Object) ([]EventHandler, error) {
	gvk, err := w.getGvk(resource)
	if err != nil {
		return nil, err
	}
	w.lock.RLock()
	handlers, ok := w.handlers[gvk]
	w.lock.RUnlock()
	if !ok {
		return nil, errors.Errorf("no handler registered for gvk %v", gvk)
	}

	return handlers, nil
}

func (w *watcher) Watch(resource runtime.Object, eventHandler EventHandler, predicates ...predicate.Predicate) error {
	// create a source for the resource type
	src := &source.Kind{Type: resource}

	// send watch events to the Cache
	if err := w.ctl.Watch(src, w.events, predicates...); err != nil {
		return err
	}

	gvk, err := w.getGvk(resource)
	if err != nil {
		return err
	}

	// add the handler to our map
	w.lock.Lock()
	// use gvk as the key
	handlers := w.handlers[gvk]
	handlers = append(handlers, eventHandler)
	w.handlers[gvk] = handlers
	w.lock.Unlock()

	return nil
}

func (w *watcher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// event key is stored in the request name
	key := request.Name
	log.Log.V(4).Info("watcher reconciling key", "key", key)

	event := w.events.Get(key)

	if event == nil {
		return reconcile.Result{}, errors.Errorf("internal error: received invalid event key %v", key)
	}

	switch event.EventType {
	case EventTypeCreate:
		obj := event.CreateEvent.Object
		handlers, err := w.getHandlers(obj)
		if err != nil {
			return reconcile.Result{}, err
		}
		for _, h := range handlers {
			err := h.Create(obj)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case EventTypeUpdate:
		obj := event.UpdateEvent.ObjectNew
		handlers, err := w.getHandlers(obj)
		if err != nil {
			return reconcile.Result{}, err
		}
		for _, h := range handlers {
			err := h.Update(event.UpdateEvent.ObjectOld, obj)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case EventTypeDelete:
		obj := event.DeleteEvent.Object
		handlers, err := w.getHandlers(obj)
		if err != nil {
			return reconcile.Result{}, err
		}
		for _, h := range handlers {
			err := h.Delete(event.DeleteEvent.Object)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case EventTypeGeneric:
		obj := event.GenericEvent.Object
		handlers, err := w.getHandlers(obj)
		if err != nil {
			return reconcile.Result{}, err
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

	w.events.Forget(key)

	return reconcile.Result{}, nil
}
