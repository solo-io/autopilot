// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	autopilot_kind_import "github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type PaintEventHandler interface {
	Create(obj *autopilot_kind_import.Paint) error
	Update(old, new *autopilot_kind_import.Paint) error
	Delete(obj *autopilot_kind_import.Paint) error
	Generic(obj *autopilot_kind_import.Paint) error
}

type PaintEventHandlerFuncs struct {
	OnCreate  func(obj *autopilot_kind_import.Paint) error
	OnUpdate  func(old, new *autopilot_kind_import.Paint) error
	OnDelete  func(obj *autopilot_kind_import.Paint) error
	OnGeneric func(obj *autopilot_kind_import.Paint) error
}

func (f *PaintEventHandlerFuncs) Create(obj *autopilot_kind_import.Paint) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PaintEventHandlerFuncs) Delete(obj *autopilot_kind_import.Paint) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PaintEventHandlerFuncs) Update(objOld, objNew *autopilot_kind_import.Paint) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PaintEventHandlerFuncs) Generic(obj *autopilot_kind_import.Paint) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type PaintController interface {
	AddEventHandler(ctx context.Context, h PaintEventHandler, predicates ...predicate.Predicate) error
}

type PaintControllerImpl struct {
	watcher events.EventWatcher
}

func NewPaintController(name string, mgr manager.Manager) (PaintController, error) {
	if err := autopilot_kind_import.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &PaintControllerImpl{
		watcher: w,
	}, nil
}

func (c *PaintControllerImpl) AddEventHandler(ctx context.Context, h PaintEventHandler, predicates ...predicate.Predicate) error {
	handler := genericPaintHandler{handler: h}
	if err := c.watcher.Watch(ctx, &autopilot_kind_import.Paint{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericPaintHandler implements a generic events.EventHandler
type genericPaintHandler struct {
	handler PaintEventHandler
}

func (h genericPaintHandler) Create(object runtime.Object) error {
	obj, ok := object.(*autopilot_kind_import.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericPaintHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*autopilot_kind_import.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericPaintHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*autopilot_kind_import.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", old)
	}
	objNew, ok := new.(*autopilot_kind_import.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericPaintHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*autopilot_kind_import.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
