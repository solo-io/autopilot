// Definitions for the Kubernetes Controllers
package controller

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1"
	"github.com/solo-io/autopilot/pkg/events"
	multicluster "github.com/solo-io/autopilot/pkg/mutlicluster"
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

func NewPaintController(mgr manager.Manager, opts events.WatcherOpts) (*PaintController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(mgr, opts)
	if err != nil {
		return nil, err
	}
	return &PaintController{
		watcher: w,
	}, nil
}

func (c *PaintController) AddEventHandler(ctx context.Context, h PaintEventHandler, predicates ...predicate.Predicate) error {
	handler := genericPaintHandler{handler: h}
	if err := c.watcher.Watch(ctx, &Paint{}, handler, predicates...); err != nil {
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

type ManagedPaintController struct {
	mgr  *multicluster.ContextualManager
	ctrl *PaintController
}

type MultiClusterPaintController struct {
	handler PaintEventHandler

	ctx         context.Context
	ctrlLock    sync.RWMutex
	controllers map[string]*ManagedPaintController
}

func (m *MultiClusterPaintController) ClusterAdded(mgr *multicluster.ContextualManager,
	name string) error {

	mcCtrl, err := m.addCluster(mgr, name, name)
	if err != nil {
		return err
	}

	m.ctrlLock.Lock()
	defer m.ctrlLock.Unlock()
	m.controllers[name] = &ManagedPaintController{
		mgr:  mcCtrl.mgr,
		ctrl: mcCtrl.ctrl,
	}
	return nil
}

func (m *MultiClusterPaintController) addCluster(mgr *multicluster.ContextualManager,
	cluster, name string) (*ManagedPaintController, error) {

	ctrl, err := NewPaintController(mgr.Manager(), events.WatcherOpts{
		Name:    name,
		Cluster: cluster,
	})
	if err != nil {
		return nil, err
	}
	if err := ctrl.AddEventHandler(m.ctx, m.handler); err != nil {
		return nil, err
	}

	return &ManagedPaintController{
		mgr:  mgr,
		ctrl: ctrl,
	}, nil
}

func (m *MultiClusterPaintController) ClusterRemoved(name string) error {
	m.ctrlLock.Lock()
	defer m.ctrlLock.Unlock()
	mgr, ok := m.controllers[name]
	if !ok {
		return eris.Errorf("could not find controller for cluster %s", name)
	}
	go mgr.mgr.Stop()
	delete(m.controllers, name)
	return nil
}

// The mgr arg here should be the local cluster manager
func NewMultiClusterPaintController(ctx context.Context, mgr manager.Manager,
	handler PaintEventHandler) (*MultiClusterPaintController, error) {

	mcCtrl := &MultiClusterPaintController{
		handler:     handler,
		ctx:         ctx,
		ctrlLock:    sync.RWMutex{},
		controllers: make(map[string]*ManagedPaintController),
	}

	ctxMgr := multicluster.NewContextualManager(ctx, mgr)
	localController, err := mcCtrl.addCluster(ctxMgr, "", "local-Paint-controller")
	if err != nil {
		return nil, err
	}
	mcCtrl.controllers[""] = localController
	return mcCtrl, nil
}
