package request

import (
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Requests wraps a slice of reconcile.Requests
type Requests []reconcile.Request

// SyncRequests is a threadsafe wrapper for
type SyncRequests struct {
	requests Requests
	lock     sync.RWMutex
}

// append a request
func (s *SyncRequests) Append(req reconcile.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.requests = append(s.requests, req)
}

// remove a request
func (s *SyncRequests) Remove(reqToDelete reconcile.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i, req := range s.requests {
		if req == reqToDelete {
			s.requests = append(s.requests[:i], s.requests[i+1:]...)
			break
		}
	}
}

// get the stored requests
func (s *SyncRequests) Requests() []reconcile.Request {
	s.lock.RLock()
	defer s.lock.RUnlock()
	out := make([]reconcile.Request, len(s.requests))
	for i := range s.requests {
		out[i] = s.requests[i]
	}
	return out
}

// convenience function for iterating over the stored requests
func (s *SyncRequests) Each(fn func(reconcile.Request)) {
	for _, req := range s.Requests() {
		fn(req)
	}
}
