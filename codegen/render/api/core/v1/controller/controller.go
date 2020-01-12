// Definitions for the Kubernetes Controllers
package controller

import (
	. "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
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

func NewSecretController(name string, mgr manager.Manager) (*SecretController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
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
