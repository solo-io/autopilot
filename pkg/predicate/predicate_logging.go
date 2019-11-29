package predicate

import (
	"fmt"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = EventLogger{}

// EventLogger logs the event
type EventLogger struct {
	Logger logr.Logger
}

func (h EventLogger) Create(e event.CreateEvent) bool {
	h.Logger.Info(fmt.Sprintf("create %T: %v.%v", e.Object, e.Meta.GetName(), e.Meta.GetNamespace()))
	return true
}

func (h EventLogger) Delete(e event.DeleteEvent) bool {
	h.Logger.Info(fmt.Sprintf("delete %T: %v.%v", e.Object, e.Meta.GetName(), e.Meta.GetNamespace()))
	return true
}

func (h EventLogger) Update(e event.UpdateEvent) bool {
	h.Logger.Info(fmt.Sprintf("update %T: %v.%v", e.ObjectNew, e.MetaNew.GetName(), e.MetaNew.GetNamespace()))
	return true
}

func (h EventLogger) Generic(e event.GenericEvent) bool {
	h.Logger.Info(fmt.Sprintf("generic %T: %v.%v", e.Object, e.Meta.GetName(), e.Meta.GetNamespace()))
	return true
}
