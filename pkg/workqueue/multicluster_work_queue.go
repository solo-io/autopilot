package workqueue

import (
	"k8s.io/client-go/util/workqueue"
	"sync"
)

// MultiClusterQueues multiplexes queues across
// multiple k8s clusters.
type MultiClusterQueues struct {
	queues map[string]workqueue.RateLimitingInterface
	lock   sync.RWMutex
}

// sets the queue for a cluster
func (s *MultiClusterQueues) Set(cluster string, queue workqueue.RateLimitingInterface) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]workqueue.RateLimitingInterface)
	}
	s.queues[cluster] = queue
}

// removes the queue for a cluster
func (s *MultiClusterQueues) Remove(cluster string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.queues, cluster)
}

// get the stored queues for the cluster
func (s *MultiClusterQueues) Get(cluster string) workqueue.RateLimitingInterface {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.queues[cluster]
}

// currently unused, useful for debugging
func (s *MultiClusterQueues) All() []workqueue.RateLimitingInterface {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var queues []workqueue.RateLimitingInterface
	for _, queue := range s.queues {
		queues = append(queues, queue)
	}
	return queues
}
