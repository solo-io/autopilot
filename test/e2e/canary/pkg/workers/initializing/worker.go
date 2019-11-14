package initializing

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/solo-io/autopilot/test/e2e/canary/pkg/parameters"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"knative.dev/pkg/network"

	"github.com/go-logr/logr"
	"github.com/solo-io/autopilot/pkg/ezkube"

	v1 "github.com/solo-io/autopilot/test/e2e/canary/pkg/apis/canarydeployments/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

type Worker struct {
	Client ezkube.Client
	Logger logr.Logger
}

func (w *Worker) Sync(ctx context.Context, canary *v1.CanaryDeployment, inputs Inputs) (Outputs, v1.CanaryDeploymentPhase, *v1.CanaryDeploymentStatusInfo, error) {

	targetDeployment, ok := inputs.FindDeployment(canary.Name, canary.Namespace)
	if !ok {
		return Outputs{}, "", nil, errors.Errorf("primary deployment not found for canary %v", canary.Name)
	}
	targetDeployment.Spec.Replicas = pointer.Int32Ptr(0)

	if targetDeployment.Spec.Template.Labels == nil {
		return Outputs{}, "", nil, errors.Errorf("invalid target deployment %v missing labels", canary.Name)
	}

	primaryName := canary.Name + "-primary"
	primaryLabels := map[string]string{
		"canary": "false",
	}

	canaryName := canary.Name + "-canary"
	canaryLabels := map[string]string{
		"canary": "true",
	}

	for k, v := range targetDeployment.Spec.Template.Labels {
		primaryLabels[k] = v
		canaryLabels[k] = v
	}

	primaryDeployment := makeDeployment(primaryName, canary.Namespace, 1, targetDeployment.Spec, targetDeployment.Annotations, primaryLabels)
	primaryService := makeService(primaryName, canary.Namespace, canary.Spec.Ports, primaryDeployment.Spec.Selector.MatchLabels)

	canaryDeployment := makeDeployment(canaryName, canary.Namespace, 0, targetDeployment.Spec, targetDeployment.Annotations, canaryLabels)
	canaryService := makeService(canaryName, canary.Namespace, canary.Spec.Ports, canaryDeployment.Spec.Selector.MatchLabels)

	// frontService is used to match requests; traffic will be split between the primary and canary
	frontService := makeService(canary.Name, canary.Namespace, canary.Spec.Ports, map[string]string{"canary": canary.Name})

	// create a route for each port
	var routes []*istiov1alpha3.HTTPRoute

	for _, port := range canary.Spec.Ports {
		routes = append(routes, &istiov1alpha3.HTTPRoute{
			Name: fmt.Sprintf("split-%v", port),
			Match: []*istiov1alpha3.HTTPMatchRequest{{
				Uri: &istiov1alpha3.StringMatch{
					MatchType: &istiov1alpha3.StringMatch_Prefix{
						Prefix: "/",
					},
				},
				Port: uint32(port),
			}},
			Route: []*istiov1alpha3.HTTPRouteDestination{
				// all traffic to primary destination
				{
					Destination: &istiov1alpha3.Destination{
						Host: network.GetServiceHostname(primaryService.Name, primaryService.Namespace),
						Port: &istiov1alpha3.PortSelector{Number: uint32(port)},
					},
					Weight: 100,
				},
				{
					Destination: &istiov1alpha3.Destination{
						Host: network.GetServiceHostname(canaryService.Name, canaryService.Namespace),
						Port: &istiov1alpha3.PortSelector{Number: uint32(port)},
					},
					Weight: 0,
				},
			},
		})
	}

	vs := v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      canary.Name,
			Namespace: canary.Namespace,
		},
		Spec: istiov1alpha3.VirtualService{
			Hosts:    []string{canary.Name},
			Gateways: []string{"mesh", "ingressgateway"}, // defaults to mesh
			Http:     routes,
		},
	}
	return Outputs{
			Deployments: parameters.Deployments{
				Items: []appsv1.Deployment{
					targetDeployment,
					primaryDeployment,
					canaryDeployment,
				},
			},
			Services: parameters.Services{
				Items: []corev1.Service{
					primaryService,
					canaryService,
					frontService,
				},
			},
			VirtualServices: parameters.VirtualServices{
				Items: []v1alpha3.VirtualService{
					vs,
				},
			},
		},
		v1.CanaryDeploymentPhaseWaiting,
		nil,
		nil
}

func makeDeployment(name, namespace string, replicas int32, fromSpec appsv1.DeploymentSpec, annotations, labels map[string]string) appsv1.Deployment {
	fromSpec.Replicas = pointer.Int32Ptr(replicas)
	fromSpec.Template.Labels = labels
	fromSpec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: fromSpec,
	}
}

func makeService(name, namespace string, ports []int32, labels map[string]string) corev1.Service {
	var kubePorts []corev1.ServicePort
	for i, p := range ports {
		kubePorts = append(kubePorts, corev1.ServicePort{
			Name: fmt.Sprintf("http-%v", i),
			Port: p,
		})
	}
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports:    kubePorts,
			Selector: labels,
		},
	}
}

func makeHosts(svcName, svcNamespace string) []string {
	return []string{
		svcName,
		svcName + "." + svcNamespace,
		svcName + "." + svcNamespace + ".svc",
		network.GetServiceHostname(svcName, svcNamespace),
	}
}
