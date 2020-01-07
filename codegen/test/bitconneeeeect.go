package test

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func RegisterDynamicReconciler(stop <-chan struct{}, name string, mgr manager.Manager, r reconcile.Reconciler, forType runtime.Object) error {
	c, err := controller.New(name, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: forType}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return c.Start(stop)
}
