package initializing

import (
	"context"
	"fmt"
	"github.com/solo-io/autopilot/examples/test/pkg/parameters"
	"istio.io/api/networking/v1alpha3"
	alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/examples/test/pkg/apis/tests/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
}

func (w *Worker) Sync(ctx context.Context, test *v1.Test, inputs Inputs) (Outputs, v1.TestPhase, *v1.TestStatusInfo, error) {
	target := test.Spec.Target
	host := target.Name + "." + target.Namespace

	var targetSvc *parameters.Service
	for _, svc := range inputs.Services.Items {
		if svc.Name == target.Name && svc.Namespace == target.Namespace {
			targetSvc = &svc
		}
	}

	if targetSvc == nil {
		return Outputs{}, "", &v1.TestStatusInfo{}, fmt.Errorf("invalid spec, service %v not found", target)
	}

	fault := (*v1alpha3.HTTPFaultInjection)(&test.Spec.Faults)

	var routes []*v1alpha3.HTTPRoute
	for _, port := range targetSvc.Spec.Ports {
		portNum := uint32(port.Port)
		routes = append(routes, &v1alpha3.HTTPRoute{
			Match: []*v1alpha3.HTTPMatchRequest{{
				Port: portNum,
			}},
			Route: []*v1alpha3.HTTPRouteDestination{{
				Destination: &v1alpha3.Destination{
					Host: host,
					Port: &v1alpha3.PortSelector{Number: portNum},
				},
			}},
			Fault: fault,
		})
	}

	return Outputs{
		VirtualServices: parameters.VirtualServices{
			Items: []alpha3.VirtualService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   test.Namespace,
						Name:        test.Name,
						Labels:      test.Labels,
						Annotations: test.Annotations,
					},
					Spec: v1alpha3.VirtualService{
						Hosts: []string{host},
						Http:  routes,
					},
				},
			}},
	}, v1.TestPhaseProcessing, &v1.TestStatusInfo{TimeStarted: metav1.Time{Time: time.Now()}}, nil
}
