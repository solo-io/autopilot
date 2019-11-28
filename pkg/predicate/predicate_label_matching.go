package handler

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = LabelMatchingPredicate{}

// LabelMatchingPredicate adds a request on a Create event and removes it on a Delete event
type LabelMatchingPredicate struct {
	Selector labels.Selector
}

func (p LabelMatchingPredicate) matches(meta v1.Object) bool {
	if meta != nil {
		return p.Selector.Matches(labels.Set(meta.GetLabels()))
	}
	return false
}

func (p LabelMatchingPredicate) Create(e event.CreateEvent) bool {
	return p.matches(e.Meta)
}

func (p LabelMatchingPredicate) Delete(e event.DeleteEvent) bool {
	return p.matches(e.Meta)
}

func (p LabelMatchingPredicate) Update(e event.UpdateEvent) bool {
	return p.matches(e.MetaOld) || p.matches(e.MetaNew)
}

func (p LabelMatchingPredicate) Generic(e event.GenericEvent) bool {
	return p.matches(e.Meta)
}
