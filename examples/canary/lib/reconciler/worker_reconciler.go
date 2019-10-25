package reconciler

import (
	"context"
	"github.com/solo-io/autopilot/examples/canary/lib/work"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
)

type Reconciler struct {
	ctx      context.Context
	kubeType runtime.Object
	access   sync.RWMutex
	mgr      manager.Manager

	runners map[reconcile.Request]work.Runner
	cancels map[reconcile.Request]func()
}

func NewReconciler(ctx context.Context,
	kubeType runtime.Object,
	mgr manager.Manager) *Reconciler {
	return &Reconciler{
		ctx:      ctx,
		kubeType: kubeType,
		mgr:      mgr,
		runners:  make(map[reconcile.Request]work.Runner),
		cancels:  make(map[reconcile.Request]func()),
	}
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	instance := r.kubeType.DeepCopyObject()
	if err := r.mgr.GetClient().Get(r.ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.deleteRunner(req)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	runner := r.getRunner(req)
	if runner == nil {
		runner = r.addRunner(r.ctx, req)
	}
}

func (r *Reconciler) getRunner(req reconcile.Request) work.Runner {
	r.access.Lock()
	defer r.access.Unlock()
	return r.runners[req]
}

func (r *Reconciler) addRunner(ctx context.Context, req reconcile.Request) work.Runner {
	worker := {}

}

func (r *Reconciler) deleteRunner(req reconcile.Request) {
	r.access.Lock()
	defer r.access.Unlock()
	delete(r.runners, req)
	cancel, ok := r.cancels[req]
	if !ok {
		return
	}
	cancel()
	delete(r.cancels, req)
}
