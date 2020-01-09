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

type PaintEventHandler struct {
	OnCreate  func(obj *Paint) error
	OnUpdate  func(old, new *Paint) error
	OnDelete  func(obj *Paint) error
	OnGeneric func(obj *Paint) error
}

func (f *PaintEventHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PaintEventHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PaintEventHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	objNew, ok := new.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PaintEventHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T")
	}
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type PaintController struct {
	watcher events.EventWatcher
}

func NewPaintController(name string, mgr manager.Manager) (*PaintController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil{
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

func (c *PaintController) AddEventHandler(h *PaintEventHandler, predicates ...predicate.Predicate) error {
	if err := c.watcher.Watch(&Paint{}, h, predicates...); err != nil {
		return err
	}
	return nil
}
