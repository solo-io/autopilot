package handler

import (
	"github.com/solo-io/autopilot/pkg/request"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ predicate.Predicate = RequestTrackingPredicate{}

// RequestTrackingPredicate adds a request on a Create event and removes it on a Delete event
type RequestTrackingPredicate struct {
	Requests *request.SyncRequests
}

func (h RequestTrackingPredicate) Create(e event.CreateEvent) bool {
	h.Requests.Append(RequestForObject(e.Meta))
	return true
}

func (h RequestTrackingPredicate) Delete(e event.DeleteEvent) bool {
	h.Requests.Remove(RequestForObject(e.Meta))
	return true
}

func (h RequestTrackingPredicate) Update(e event.UpdateEvent) bool {
	return true
}

func (h RequestTrackingPredicate) Generic(e event.GenericEvent) bool {
	return true
}

func RequestForObject(meta v1.Object) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}}
}
