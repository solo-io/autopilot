package handler

import (
	"github.com/solo-io/autopilot/pkg/request"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ handler.EventHandler = &MultiClusterRequestTracker{}

// MultiClusterRequestTracker tracks reconcile requests across clusters
// It is used to map requests for input resources (in any cluster)
// back to the original
type MultiClusterRequestTracker struct {
	Cluster  string
	Requests *request.MultiClusterRequests
}

func (h *MultiClusterRequestTracker) Create(evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
	h.Requests.Append(h.Cluster, RequestForObject(evt.Meta))
}

func (h *MultiClusterRequestTracker) Delete(evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	h.Requests.Remove(h.Cluster, RequestForObject(evt.Meta))
}

func (h *MultiClusterRequestTracker) Update(event.UpdateEvent, workqueue.RateLimitingInterface) {}

func (h *MultiClusterRequestTracker) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {}

func RequestForObject(meta v1.Object) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}}
}
