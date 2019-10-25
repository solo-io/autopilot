package canary

import (
	"context"
	"github.com/solo-io/autopilot/examples/canary/lib/reconciler"
	glookubev1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/apis/gloo.solo.io/v1"
	v1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	canariesv1 "github.com/solo-io/autopilot/examples/canary/pkg/apis/canaries/v1"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("controller_canary")

// Add creates a new Canary Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return reconciler.NewReconciler(context.TODO(), &canariesv1.Canary{}, mgr)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("canary-controller").
		For(&canariesv1.Canary{}).
		Owns(&v1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&glookubev1.UpstreamGroup{}).
		Complete(r)
}
