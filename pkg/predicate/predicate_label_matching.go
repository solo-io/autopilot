package predicate

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = LabelMatcher{}

// LabelMatcher filters events by the event's object labels
type LabelMatcher struct {
	Selector labels.Selector
}

func (p LabelMatcher) matches(meta v1.Object) bool {
	if meta != nil {
		return p.Selector.Matches(labels.Set(meta.GetLabels()))
	}
	return false
}

func (p LabelMatcher) Create(e event.CreateEvent) bool {
	return p.matches(e.Meta)
}

func (p LabelMatcher) Delete(e event.DeleteEvent) bool {
	return p.matches(e.Meta)
}

func (p LabelMatcher) Update(e event.UpdateEvent) bool {
	return p.matches(e.MetaOld) || p.matches(e.MetaNew)
}

func (p LabelMatcher) Generic(e event.GenericEvent) bool {
	return p.matches(e.Meta)
}
