// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	things_test_io_v1 "github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Handle events for the Paint Resource
type PaintCache interface {
	List() ([]*things_test_io_v1.Paint, error)
}

type paintCache struct {
	client manager.Manager
}

func (p *paintCache) CreatePaint(obj *things_test_io_v1.Paint) error {
	p.mgr.GetCache().List()
	return nil
}

func (p *paintCache) UpdatePaint(old, new *things_test_io_v1.Paint) error {
	return nil
}

func (p *paintCache) DeletePaint(obj *things_test_io_v1.Paint) error {
	return nil
}

func (p *paintCache) GenericPaint(obj *things_test_io_v1.Paint) error {
	return nil
}

func (f *PaintEventHandlerFuncs) CreatePaint(obj *things_test_io_v1.Paint) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PaintEventHandlerFuncs) DeletePaint(obj *things_test_io_v1.Paint) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PaintEventHandlerFuncs) UpdatePaint(objOld, objNew *things_test_io_v1.Paint) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PaintEventHandlerFuncs) GenericPaint(obj *things_test_io_v1.Paint) error {
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
	if err := things_test_io_v1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	return &PaintControllerImpl{
		watcher: events.NewWatcher(name, mgr, &things_test_io_v1.Paint{}),
	}, nil
}

func (c *PaintControllerImpl) AddEventHandler(ctx context.Context, h PaintEventHandler, predicates ...predicate.Predicate) error {
	handler := genericPaintHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericPaintHandler implements a generic events.EventHandler
type genericPaintHandler struct {
	handler PaintEventHandler
}

func (h genericPaintHandler) Create(object runtime.Object) error {
	obj, ok := object.(*things_test_io_v1.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return h.handler.CreatePaint(obj)
}

func (h genericPaintHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*things_test_io_v1.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return h.handler.DeletePaint(obj)
}

func (h genericPaintHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*things_test_io_v1.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", old)
	}
	objNew, ok := new.(*things_test_io_v1.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", new)
	}
	return h.handler.UpdatePaint(objOld, objNew)
}

func (h genericPaintHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*things_test_io_v1.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return h.handler.GenericPaint(obj)
}
