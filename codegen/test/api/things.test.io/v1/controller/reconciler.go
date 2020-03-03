// Definitions for the Kubernetes Controllers
package controller

import (
	"context"
	"time"

	things_test_io_v1 "github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile the Paint Resource
// implemented by the user
type PaintReconciler interface {
	Reconcile(obj *things_test_io_v1.Paint) (time.Duration, error)
}

// Reconcile and finalize the Paint Resource
// implemented by the user
type FinalizingPaintReconciler interface {
	PaintReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	FinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	Finalize(obj *things_test_io_v1.Paint) error
}

type PaintReconcileLoop interface {
	RunReconciler(ctx context.Context, rec PaintReconciler, predicates ...predicate.Predicate) error
}

type PaintReconcileLoopImpl struct {
	loop reconcile.Loop
}

func NewPaintReconcileLoopImpl(name string, mgr manager.Manager) (PaintReconcileLoop, error) {
	if err := things_test_io_v1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	return &PaintReconcileLoopImpl{
		loop: reconcile.NewLoop(name, mgr, &things_test_io_v1.Paint{}),
	}, nil
}

func (c *PaintReconcileLoopImpl) RunReconciler(ctx context.Context, reconciler PaintReconciler, predicates ...predicate.Predicate) error {
	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(FinalizingPaintReconciler); ok {
		reconcilerWrapper = genericFinalizingPaintReconciler{
			finalizingReconciler: finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericPaintReconciler{
			reconciler: reconciler,
		}
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericPaintHandler implements a generic reconcile.Reconciler
type genericPaintReconciler struct {
	reconciler PaintReconciler
}

func (r genericPaintReconciler) Reconcile(object ezkube.Object) (time.Duration, error) {
	obj, ok := object.(*things_test_io_v1.Paint)
	if !ok {
		return 0, errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return r.reconciler.Reconcile(obj)
}

// genericFinalizingPaintReconciler implements a generic reconcile.FinalizingReconciler
type genericFinalizingPaintReconciler struct {
	finalizingReconciler FinalizingPaintReconciler
}

func (r genericFinalizingPaintReconciler) Reconcile(object ezkube.Object) (time.Duration, error) {
	rec := genericPaintReconciler{reconciler: r.finalizingReconciler}
	return rec.Reconcile(object)
}

func (r genericFinalizingPaintReconciler) FinalizerName() string {
	return r.finalizingReconciler.FinalizerName()
}

func (r genericFinalizingPaintReconciler) Finalize(object ezkube.Object) error {
	obj, ok := object.(*things_test_io_v1.Paint)
	if !ok {
		return errors.Errorf("internal error: Paint handler received event for %T", object)
	}
	return r.finalizingReconciler.Finalize(obj)
}
