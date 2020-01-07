package reconcilefunc

import (
	"context"
	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/controller"
	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Run(ctx context.Context) error {
	ctl := controller.Controller{
		Name: "test",
		Ctx:  ctx,
		Reconcile: func(primaryResource runtime.Object) (result reconcile.Result, e error) {
			return recFunc(primaryResource)
		},
		PrimaryResource:   &v1.CanaryDeployment{},
		PrimaryPredicates: predicates,
		InputResources: map[runtime.Object][]predicate.Predicate{
			&v1.Secret{}: predicates,
		},
	}
}

type CanaryReconciler struct {
}

func (r *CanaryReconciler) Reconcile(primaryResource runtime.Object) (reconcile.Result, error) {
	canaryDeployment, ok := primaryResource.(*v1.CanaryDeployment)
	if !ok {
		return reconcile.Result{}, errors.Errorf("invalid type for canary reconciler: %T", primaryResource)
	}
	switch canaryDeployment.Status.Phase {

	}
}

type CanaryHandler interface {
	HandleCanaryInitializing()
	HandleCanaryWaiting()
}
