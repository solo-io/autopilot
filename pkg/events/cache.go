package events

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Go's "union type"
type eventType interface {
	meta() metav1.Object
	isEvent()
}

type createEvent event.CreateEvent

func (e createEvent) meta() metav1.Object {
	return e.Meta
}
func (e createEvent) isEvent() {}

type updateEvent event.UpdateEvent

func (e updateEvent) meta() metav1.Object {
	return e.MetaNew
}
func (e updateEvent) isEvent() {}

type deleteEvent event.DeleteEvent

func (e deleteEvent) meta() metav1.Object {
	return e.Meta
}
func (e deleteEvent) isEvent() {}

type genericEvent event.GenericEvent

func (e genericEvent) meta() metav1.Object {
	return e.Meta
}
func (e genericEvent) isEvent() {}

// Cache caches k8s resource events
// It implements handler.eventHandler,
// emitting reconcile Requests for its cached
// events. This allows a Reconciler to
// claim and process these custom events
type Cache interface {
	// handler that receives events from controller-runtime
	handler.EventHandler

	// retrieve an event from the cache
	Get(key string) eventType

	// remove an event from the cache
	Forget(key string)
}

type cache struct {
	lock sync.RWMutex

	// cache keys will be mapped to the reconcile.Request.Name
	cache map[string]eventType
}

func NewCache() *cache {
	return &cache{cache: make(map[string]eventType)}
}

func (c *cache) handleEvent(evt eventType, q workqueue.RateLimitingInterface) {

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
	c.handleEvent(createEvent(evt), q)
}

func (c *cache) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(updateEvent(evt), q)
}

func (c *cache) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(deleteEvent(evt), q)
}

func (c *cache) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	c.handleEvent(genericEvent(evt), q)
}

// pops the event with the key
func (c *cache) Get(key string) eventType {
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
