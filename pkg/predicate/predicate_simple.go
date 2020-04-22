package predicate

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = SimplePredicate{}

// SimpleEventFilter filters events for a single object type
type SimpleEventFilter interface {
	// return True to filter out the event
	FilterEvent(obj v1.Object) bool
}

// SimplePredicate filters events based on a ShouldSync function
type SimplePredicate struct {
	Filter SimpleEventFilter
}

func (p SimplePredicate) Create(e event.CreateEvent) bool {
	return !p.Filter.FilterEvent(e.Meta)
}

func (p SimplePredicate) Delete(e event.DeleteEvent) bool {
	return !p.Filter.FilterEvent(e.Meta)
}

func (p SimplePredicate) Update(e event.UpdateEvent) bool {
	return !p.Filter.FilterEvent(e.MetaNew)
}

func (p SimplePredicate) Generic(e event.GenericEvent) bool {
	return !p.Filter.FilterEvent(e.Meta)
}
