// Definitions for the Kubernetes Controllers
package controller

import (
	. "github.com/solo-io/autopilot/codegen/render/api/things.test.io/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type PaintEventHandler interface {
	Create(obj *Paint) error
	Update(old, new *Paint) error
	Delete(obj *Paint) error
	Generic(obj *Paint) error
}

type PaintEventHandlerFuncs struct {
	OnCreate  func(obj *Paint) error
	OnUpdate  func(old, new *Paint) error
	OnDelete  func(obj *Paint) error
	OnGeneric func(obj *Paint) error
}

func (f *PaintEventHandlerFuncs) Create(obj *Paint) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PaintEventHandlerFuncs) Delete(obj *Paint) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PaintEventHandlerFuncs) Update(objOld, objNew *Paint) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PaintEventHandlerFuncs) Generic(obj *Paint) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type PaintController struct {
	watcher events.EventWatcher
}

func NewPaintController(name string, mgr manager.Manager) (*PaintController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &PaintController{
		watcher: w,
	}, nil
}

func (c *PaintController) AddEventHandler(h PaintEventHandler, predicates ...predicate.Predicate) error {
	handler := genericPaintHandler{handler: h}
	if err := c.watcher.Watch(&Paint{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericPaintHandler implements a generic events.EventHandler
type genericPaintHandler struct {
	handler PaintEventHandler
}

func (h genericPaintHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	return h.handler.Create(obj)
}

func (h genericPaintHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	return h.handler.Delete(obj)
}

func (h genericPaintHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	objNew, ok := new.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericPaintHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	return h.handler.Generic(obj)
}
