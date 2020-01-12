package source

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Stoppable is a stoppable source
type Stoppable interface {
	source.Source
	inject.Stoppable
}

// DynamicSource is a funnel for sources that can be
// dynamically (de)registered before & after the controller has started
type Dynamic interface {
	source.Source

	// sources must be registered with a unique id
	Add(id string, src source.Source) error

	// remove a source. errors if not found
	Remove(id string) error
}

// cache of sources
type cachedSource struct {
	// the original source
	source source.Source

	// cancel function to stop it
	cancel context.CancelFunc
}

// the args with which the dynamic source was started
type startArgs struct {
	h  handler.EventHandler
	i  workqueue.RateLimitingInterface
	ps []predicate.Predicate
}

// DynamicSource implements Dynamic
type DynamicSource struct {
	// cancel this context to stop all registered sources
	ctx context.Context

	// the cached sources that can be dynamically added/removed
	cache map[string]cachedSource

	// cache access
	lock sync.RWMutex

	// has source started?
	started *startArgs

	// the channel to which to push events
	output source.Channel
}

func NewDynamicSource(ctx context.Context) *DynamicSource {
	return &DynamicSource{
		ctx:   ctx,
		cache: make(map[string]cachedSource),
	}
}

// start all the sources
func (s *DynamicSource) Start(h handler.EventHandler, i workqueue.RateLimitingInterface, ps ...predicate.Predicate) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.started != nil {
		return errors.Errorf("source was already started")
	}

	for _, src := range s.cache {

		if err := src.source.Start(h, i, ps...); err != nil {
			return err
		}
	}

	s.started = &startArgs{
		h:  h,
		i:  i,
		ps: ps,
	}

	return nil
}

// only Stoppable sources are currently supported
func (s *DynamicSource) Add(id string, src Stoppable) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, exists := s.cache[id]; exists {
		return errors.Errorf("source %v already exists")
	}

	ctx, cancel := context.WithCancel(s.ctx)
	if err := src.InjectStopChannel(ctx.Done()); err != nil {
		return err
	}

	if s.started != nil {
		if err := src.Start(s.started.h, s.started.i, s.started.ps...); err != nil {
			return errors.Wrapf(err, "failed to start source %v", id)
		}
	}

	s.cache[id] = cachedSource{
		source: src,
		cancel: cancel,
	}

	return nil
}

// remove (and stop) a source
func (s *DynamicSource) Remove(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	src, ok := s.cache[id]
	if !ok {
		return errors.Errorf("no source in cache with id %v", id)
	}

	src.cancel()

	delete(s.cache, id)

	return nil
}
