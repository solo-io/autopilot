// NOTE: Boilerplate only.  Ignore this file.

// GoPackage v1 contains API Schema definitions for the canaries v1 API group
// +k8s:deepcopy-gen=package,register
// +groupName=internal.autopilot.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme

// AddToScheme adds all Resources to the Scheme

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "internal.autopilot.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
