package handler

import (
	"github.com/solo-io/autopilot/pkg/request"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var enqueueStaticLog = log.Log.WithName("eventhandler").WithName("EnqueueStaticRequests")

var _ handler.EventHandler = &EnqueueStaticRequests{}

// EnqueueStaticRequests enqueues a statically defined Request.
// This is used by Autopilot to enqueue requests for a top level resource
// whenever a watched input resource changes.
type EnqueueStaticRequests struct {
	RequestsToEnqueue *request.SyncRequests
}

// Create implements EventHandler
func (e *EnqueueStaticRequests) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueStaticLog.Error(nil, "CreateEvent received with no metadata", "event", evt)
		return
	}
	e.enqueue(q)
}

// Update implements EventHandler
func (e *EnqueueStaticRequests) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if evt.MetaOld != nil {
		e.enqueue(q)
	} else {
		enqueueStaticLog.Error(nil, "UpdateEvent received with no old metadata", "event", evt)
	}

	if evt.MetaNew != nil {
		e.enqueue(q)
	} else {
		enqueueStaticLog.Error(nil, "UpdateEvent received with no new metadata", "event", evt)
	}
}

// Delete implements EventHandler
func (e *EnqueueStaticRequests) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueStaticLog.Error(nil, "DeleteEvent received with no metadata", "event", evt)
		return
	}
	e.enqueue(q)
}

// Generic implements EventHandler
func (e *EnqueueStaticRequests) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueStaticLog.Error(nil, "GenericEvent received with no metadata", "event", evt)
		return
	}
	e.enqueue(q)
}

func (e *EnqueueStaticRequests) enqueue(q workqueue.RateLimitingInterface) {
	e.RequestsToEnqueue.Each(func(i reconcile.Request) {
		q.Add(i)
	})
}
