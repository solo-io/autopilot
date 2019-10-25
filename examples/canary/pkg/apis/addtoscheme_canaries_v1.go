package apis

import (
	v1 "github.com/solo-io/autopilot/examples/canary/pkg/apis/canaries/v1"

	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/client/clientset/versioned/scheme"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1.SchemeBuilder.AddToScheme)

	// TODO make this automatic
	AddToSchemes = append(AddToSchemes, scheme.AddToScheme)
}
