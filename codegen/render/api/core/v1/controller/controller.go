// Definitions for the Kubernetes Controllers
package controller

import (
	"context"
	"sync"

	. "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"github.com/solo-io/autopilot/pkg/events"
	multicluster "github.com/solo-io/autopilot/pkg/mutlicluster"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretEventHandler interface {
	Create(obj *Secret) error
	Update(old, new *Secret) error
	Delete(obj *Secret) error
	Generic(obj *Secret) error
}

type SecretEventHandlerFuncs struct {
	OnCreate  func(obj *Secret) error
	OnUpdate  func(old, new *Secret) error
	OnDelete  func(obj *Secret) error
	OnGeneric func(obj *Secret) error
}

func (f *SecretEventHandlerFuncs) Create(obj *Secret) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *SecretEventHandlerFuncs) Delete(obj *Secret) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *SecretEventHandlerFuncs) Update(objOld, objNew *Secret) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *SecretEventHandlerFuncs) Generic(obj *Secret) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type SecretController struct {
	watcher events.EventWatcher
}

func NewSecretController(mgr manager.Manager, opts events.WatcherOpts) (*SecretController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(mgr, opts)
	if err != nil {
		return nil, err
	}
	return &SecretController{
		watcher: w,
	}, nil
}

func (c *SecretController) AddEventHandler(h SecretEventHandler, predicates ...predicate.Predicate) error {
	handler := genericSecretHandler{handler: h}
	if err := c.watcher.Watch(&Secret{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericSecretHandler implements a generic events.EventHandler
type genericSecretHandler struct {
	handler SecretEventHandler
}

func (h genericSecretHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T")
	}
	return h.handler.Create(obj)
}

func (h genericSecretHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T")
	}
	return h.handler.Delete(obj)
}

func (h genericSecretHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T")
	}
	objNew, ok := new.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T")
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericSecretHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T")
	}
	return h.handler.Generic(obj)
}

type ManagedSecretController struct {
	mgr  *multicluster.ContextualManager
	ctrl *SecretController
}

type MultiClusterSecretController struct {
	handler SecretEventHandler

	ctx         context.Context
	ctrlLock    sync.RWMutex
	controllers map[string]*ManagedSecretController
}

func (m *MultiClusterSecretController) ClusterAdded(mgr *multicluster.ContextualManager,
	name string) error {

	mcCtrl, err := m.addCluster(mgr, name)
	if err != nil {
		return err
	}

	m.ctrlLock.Lock()
	defer m.ctrlLock.Unlock()
	m.controllers[name] = &ManagedSecretController{
		mgr:  mcCtrl.mgr,
		ctrl: mcCtrl.ctrl,
	}
	return nil
}

func (m *MultiClusterSecretController) addCluster(mgr *multicluster.ContextualManager,
	name string) (*ManagedSecretController, error) {

	ctrl, err := NewSecretController(mgr.Manager(), events.WatcherOpts{
		Name:    name,
		Cluster: name,
	})
	if err != nil {
		return nil, err
	}
	if err := ctrl.AddEventHandler(m.handler); err != nil {
		return nil, err
	}

	return &ManagedSecretController{
		mgr:  mgr,
		ctrl: ctrl,
	}, nil
}

func (m *MultiClusterSecretController) ClusterRemoved(name string) error {
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

func NewMultiClusterSecretController(ctx context.Context, mgr manager.Manager,
	handler SecretEventHandler) (*MultiClusterSecretController, error) {

	mcCtrl := &MultiClusterSecretController{
		handler:     handler,
		ctx:         ctx,
		ctrlLock:    sync.RWMutex{},
		controllers: make(map[string]*ManagedSecretController),
	}
	ctxMgr := &multicluster.ContextualManager{
		Manager: mgr,
		Context: ctx,
	}
	return mcCtrl, nil
}
