package handler

import (
	apqueue "github.com/solo-io/autopilot/pkg/workqueue"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sync"
)

// QueueRegisteringHandler registers the queue on the first create event
// it receives to a multi cluster queue registry.
func QueueRegisteringHandler(cluster string, queues *apqueue.MultiClusterQueues) handler.EventHandler {
	do := &sync.Once{}
	return &handler.Funcs{
		CreateFunc: func(_ event.CreateEvent, queue workqueue.RateLimitingInterface) {
			do.Do(func() {
				queues.Set(cluster, queue)
			})
		},
	}
}

