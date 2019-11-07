// the parameters package makes it easy to interact with input/output types for your phases
// it also handles registering the types with the kubernetes runtime.Scheme
package parameters

import (
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"

	"github.com/solo-io/autopilot/pkg/run"
)

type (

	// type aliases for corev1.Service
	Service  = corev1.Service
	Services = corev1.ServiceList

	// type aliases for istiov1alpha3.VirtualService
	VirtualService  = istiov1alpha3.VirtualService
	VirtualServices = istiov1alpha3.VirtualServiceList

	// type aliases for istiov1alpha3.Gateway
	Gateway  = istiov1alpha3.Gateway
	Gateways = istiov1alpha3.GatewayList
)

func init() {
	run.RegisterAddToScheme(istiov1alpha3.AddToScheme)
}
