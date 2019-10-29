package work

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sync"
	"time"
)

type Worker interface{
	IsWorker() // no generics > generate!
}

type Runner interface {
	Start()
	GetInterval() time.Duration
	SetInstance(instance runtime.Object)
}

type Instance []reflect.Value

// Runner holds the reference to a work schedule
type workRunner struct {
	ctx            context.Context
	doWorkMethod   reflect.Value
	teardownMethod reflect.Value
	instanceAccess sync.RWMutex
	instance       Instance
	done           <-chan struct{}
	ticker         *time.Ticker
	interval       time.Duration
}

func NewRunner(ctx context.Context, w Worker, interval time.Duration) *workRunner {
	doWorkMethod := reflect.ValueOf(w).MethodByName("DoWork")
	if !doWorkMethod.IsValid() {
		panic(fmt.Sprintf("cannot use %T as Worker, must implement DoWork", w))
	}

	teardownMethod := reflect.ValueOf(w).MethodByName("Teardown")
	if !teardownMethod.IsValid() {
		panic(fmt.Sprintf("cannot use %T as Worker, must implement Teardown", w))
	}

	return &workRunner{
		ctx:            ctx,
		doWorkMethod:   doWorkMethod,
		teardownMethod: teardownMethod,
		interval:       interval,
		done:           ctx.Done(),
	}
}

// Start runs the canary analysis on a schedule
func (w *workRunner) Start() {
	w.ticker = time.NewTicker(w.interval)
	go func() {
		// run the work function creation
		w.doWorkMethod.Call(w.instance)
		for {
			select {
			case <-w.ticker.C:
				w.doWorkMethod.Call(w.instance)
			case <-w.done:
				w.teardownMethod.Call(w.instance)
				w.ticker.Stop()
				return
			}
		}
	}()
}

func (w *workRunner) GetInterval() time.Duration {
	return w.interval
}

func (w *workRunner) GetInstance() Instance {
	w.instanceAccess.RLock()
	defer w.instanceAccess.RUnlock()
	return w.instance
}

func (w *workRunner) SetInstance(instance runtime.Object) {
	w.instanceAccess.Lock()
	w.instance = Instance{reflect.ValueOf(w.ctx), reflect.ValueOf(instance)}
	w.instanceAccess.Unlock()
}
