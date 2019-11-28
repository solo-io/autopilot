package handler

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// MultiHandler wraps and calls multiple event handlers as a single handler.Handler
type MultiHandler struct {
	Handlers []handler.EventHandler
}

func (h *MultiHandler) Create(evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
	for _, hl := range h.Handlers {
		hl.Create(evt, queue)
	}
}

func (h *MultiHandler) Update(evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	for _, hl := range h.Handlers {
		hl.Update(evt, queue)
	}
}

func (h *MultiHandler) Delete(evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	for _, hl := range h.Handlers {
		hl.Delete(evt, queue)
	}
}

func (h *MultiHandler) Generic(evt event.GenericEvent, queue workqueue.RateLimitingInterface) {
	for _, hl := range h.Handlers {
		hl.Generic(evt, queue)
	}
}
