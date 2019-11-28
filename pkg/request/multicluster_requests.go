package request

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
)

// MultiClusterRequests multiplexes reconcile Requests
// for multiple k8s clusters.
type MultiClusterRequests struct {
	requests map[string]*SyncRequests
	lock     sync.RWMutex
}

// append a request to a cluster
func (s *MultiClusterRequests) Append(cluster string, req reconcile.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.requests == nil {
		s.requests = make(map[string]*SyncRequests)
	}
	clusterRequests := s.requests[cluster]
	if clusterRequests == nil {
		clusterRequests = &SyncRequests{}
	}
	clusterRequests.Append(req)
	s.requests[cluster] = clusterRequests
}

// remove a request from a cluster
func (s *MultiClusterRequests) Remove(cluster string, reqToDelete reconcile.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.requests[cluster].Remove(reqToDelete)
}

// get the stored requests for the cluster
func (s *MultiClusterRequests) Requests(cluster string) []reconcile.Request {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.requests[cluster].Requests()
}

// convenience function for iterating over the stored requests, by cluster
func (s *MultiClusterRequests) Each(fn func(string, reconcile.Request)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for cluster := range s.requests {
		for _, req := range s.Requests(cluster) {
			fn(cluster, req)
		}
	}
}
