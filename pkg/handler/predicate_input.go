package handler

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// DependencyFilter filters events for a given top-level request before enqueuing the keys.
type DependencyFilter interface {
	// Create returns true if the Create event should be processed for the given top-level request
	Create(reconcile.Request, event.CreateEvent) bool

	// Delete returns true if the Delete event should be processed for the given top-level request
	Delete(reconcile.Request, event.DeleteEvent) bool

	// Update returns true if the Update event should be processed for the given top-level request
	Update(reconcile.Request, event.UpdateEvent) bool

	// Generic returns true if the Generic event should be processed for the given top-level request
	Generic(reconcile.Request, event.GenericEvent) bool
}

var _ predicate.Predicate = DependencyFilterPredicate{}

// DependencyPredicate provides a means to filter queueing requests for a parent object
// when an event for a dependency (autopilot input) is received
type DependencyFilterPredicate struct {
	TopLevelRequest reconcile.Request
	Filter          DependencyFilter
}

func (h DependencyFilterPredicate) Create(e event.CreateEvent) bool {
	return h.Filter.Create(h.TopLevelRequest, e)
}

func (h DependencyFilterPredicate) Delete(e event.DeleteEvent) bool {
	return h.Filter.Delete(h.TopLevelRequest, e)
}

func (h DependencyFilterPredicate) Update(e event.UpdateEvent) bool {
	return h.Filter.Update(h.TopLevelRequest, e)
}

func (h DependencyFilterPredicate) Generic(e event.GenericEvent) bool {
	return h.Filter.Generic(h.TopLevelRequest, e)
}
