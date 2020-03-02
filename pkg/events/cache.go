package events

import (
	"sync"

	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EventType string

const (
	EventTypeCreate  EventType = "Create"
	EventTypeUpdate  EventType = "Update"
	EventTypeDelete  EventType = "Delete"
	EventTypeGeneric EventType = "Generic"
)

// event represents a standard kubernetes event
// it can be
type Event struct {
	EventType EventType

	// only one can be set
	CreateEvent  *event.CreateEvent
	UpdateEvent  *event.UpdateEvent
	DeleteEvent  *event.DeleteEvent
	GenericEvent *event.GenericEvent
}

// Cache caches k8s resource events
// It implements handler.EventHandler,
// emitting reconcile Requests for its cached
// events. This allows a Reconciler to
// claim and process these custom events
type Cache interface {
	// handler that receives events from controller-runtime
	handler.EventHandler

	// retrieve an event from the cache
	Get(key string) *Event

	// remove an event from the cache
	Forget(key string)
}

type cache struct {
	lock sync.RWMutex

	// cache keys will be mapped to the reconcile.Request.Name
	cache map[string]*Event
}

func NewCache() *cache {
	return &cache{cache: make(map[string]*Event)}
}

func (c *cache) handleEvent(evt *Event, q workqueue.RateLimitingInterface) {

	// use a UUID so the reconciler can claim this event with
	// the reconcile request
	key := uuid.New()

	log.Log.V(1).Info("storing event", "key", key, "event", evt)
	c.lock.Lock()
	c.cache[key] = evt
	c.lock.Unlock()

	// add a request the event to the queue
	// the controller will Pop() it when the request reaches the reconcile function
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      key,
		Namespace: "", // no namespace required
	}})
}

func (c *cache) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(&Event{EventType: EventTypeCreate, CreateEvent: &evt}, q)
}

func (c *cache) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(&Event{EventType: EventTypeUpdate, UpdateEvent: &evt}, q)
}

func (c *cache) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(&Event{EventType: EventTypeDelete, DeleteEvent: &evt}, q)
}

func (c *cache) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(&Event{EventType: EventTypeGeneric, GenericEvent: &evt}, q)
}

// pops the event with the key
func (c *cache) Get(key string) *Event {
	c.lock.RLock()
	evt := c.cache[key]
	c.lock.RUnlock()
	return evt
}

func (c *cache) Forget(key string) {
	c.lock.Lock()
	delete(c.cache, key)
	c.lock.Unlock()
}
