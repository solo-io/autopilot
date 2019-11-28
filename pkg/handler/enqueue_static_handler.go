package handler

import (
	"github.com/solo-io/autopilot/pkg/request"
	apqueue "github.com/solo-io/autopilot/pkg/workqueue"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var enqueueMultiClusterLog = log.Log.WithName("eventhandler").WithName("EnqueueMultiCluster")

var _ handler.EventHandler = &EnqueueMultiCluster{}

// EnqueueMultiCluster enqueues statically defined requests across clusters
// whenever an event is received. Use this to propagate a list of requests
// to queues shared across cluster managers.
// This is used by Autopilot to enqueueRequestsAllClusters requests for a primary level resource
// whenever a watched input resource changes, regardless of the cluster the primary resource lives in.
type EnqueueMultiCluster struct {
	// the set of all requests to enqueueRequestsAllClusters by the target cluster (where the primary resource lives)
	RequestsToEnqueue *request.MultiClusterRequests

	// use this to queue requests to controllers registered to another manager
	WorkQueues *apqueue.MultiClusterQueues
}

// Create implements EventHandler
func (e *EnqueueMultiCluster) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueMultiClusterLog.Error(nil, "CreateEvent received with no metadata", "event", evt)
		return
	}
	e.enqueueRequestsAllClusters()
}

// Update implements EventHandler
func (e *EnqueueMultiCluster) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if evt.MetaOld != nil {
		e.enqueueRequestsAllClusters()
	} else {
		enqueueMultiClusterLog.Error(nil, "UpdateEvent received with no old metadata", "event", evt)
	}

	if evt.MetaNew != nil {
		e.enqueueRequestsAllClusters()
	} else {
		enqueueMultiClusterLog.Error(nil, "UpdateEvent received with no new metadata", "event", evt)
	}
}

// Delete implements EventHandler
func (e *EnqueueMultiCluster) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueMultiClusterLog.Error(nil, "DeleteEvent received with no metadata", "event", evt)
		return
	}
	e.enqueueRequestsAllClusters()
}

// Generic implements EventHandler
func (e *EnqueueMultiCluster) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueMultiClusterLog.Error(nil, "GenericEvent received with no metadata", "event", evt)
		return
	}
	e.enqueueRequestsAllClusters()
}

//
func (e *EnqueueMultiCluster) enqueueRequestsAllClusters() {
	e.RequestsToEnqueue.Each(func(cluster string, i reconcile.Request) {
		q := e.WorkQueues.Get(cluster)
		if q == nil {
			enqueueMultiClusterLog.Error(nil, "Cannot enqueue request, no queue registered for cluster", "request", i, "cluster", cluster)
			return
		}
		q.Add(i)
	})
}
