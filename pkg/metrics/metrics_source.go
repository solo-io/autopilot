package metrics

import (
	"time"

	v1 "github.com/solo-io/autopilot/pkg/internal/apis/metrics/v1"
	"github.com/solo-io/autopilot/pkg/internal/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logger = log.Log.WithName("metrics_event_source")

// a query func is a lambda that encapsulates a query request.
type QueryFunc func() (*QueryResult, error)

// A source of Generic controller-runtime events
// that can be plugged into a Manager/Controller
type EventSource struct {
	// the EventSource calls this func on the given interval
	QueryFunc QueryFunc

	// the interval over which to poll the metric
	Interval time.Duration

	// the logical name of the metric query being run
	// should be unique across metrics queries for a controller
	Name string

	// close this channel to kill the source
	stop <-chan struct{}
}

func (s *EventSource) Start(h handler.EventHandler, q workqueue.RateLimitingInterface, p ...predicate.Predicate) error {
	out := source.Channel{
		Source: s.events(),
	}
	return out.Start(h, q, p...)
}

func (s *EventSource) InjectStopChannel(stop <-chan struct{}) error {
	s.stop = stop
	return nil
}

func (s *EventSource) events() <-chan event.GenericEvent {
	events := make(chan event.GenericEvent, 10)

	go func() {
		interval := s.Interval

		if interval == 0 {
			interval = time.Second * 5
		}

		tick := time.Tick(interval)
		for {

			select {
			case <-tick:
				result, err := s.QueryFunc()
				if err != nil {
					logger.Error(err, "running query function failed")
					continue
				}

				metricsResult := &v1.MetricValue{
					ObjectMeta: metav1.ObjectMeta{
						Name:      s.Name,
						Namespace: constants.InternalNamespace,
					},
					Value: &v1.Value{
						Value: result,
					},
				}

				metricsEvent := event.GenericEvent{
					Meta:   metricsResult,
					Object: metricsResult,
				}

				select {
				case events <- metricsEvent:
				case <-s.stop:
					return
				}
			case <-s.stop:
				close(events)
				return
			}
		}
	}()

	return events
}
