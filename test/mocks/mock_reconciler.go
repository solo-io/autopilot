package mocks

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MockReconciler struct {
	Name    string
	Ctx     context.Context
	Cancel  context.CancelFunc
	RecFunc func(reconcile.Request) (reconcile.Result, error)
}

func NewMockReconciler(name string) *MockReconciler {
	ctx, cancel := context.WithCancel(context.Background())
	return &MockReconciler{Name: name, Ctx: ctx, Cancel: cancel}
}

func (r *MockReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	return r.RecFunc(req)
}
