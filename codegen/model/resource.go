package model

import (
	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Group struct {
	// the group version of the group
	schema.GroupVersion

	// the go  module this group belongs to
	Module string

	// the root directory for generated API code
	ApiRoot string

	// search protos recursively starting from this directory.
	// will default vendor_any if empty
	ProtoDir string

	// the kinds in the group
	Resources []Resource

	// Should we compile protos?
	RenderProtos bool

	// Should we generate kubernetes manifests?
	RenderManifests bool

	// Should we generate kubernetes Go structs?
	RenderTypes bool

	// Should we generate kubernetes Go clients?
	RenderClients bool

	// Should we generate kubernetes Go controllers?
	RenderController bool

	// custom import path to the package
	// containing the Go types
	// use this if you are generating controllers
	// for types in an external project
	CustomTypesImportPath string

	// proto descriptors will be available to the templates if the group was compiled with them.
	Descriptors []*model.DescriptorWithPath
}

// ensures the resources point to this group
func (g *Group) Init() {
	for i, resource := range g.Resources {
		resource.Group = *g
		g.Resources[i] = resource
	}
}

type Resource struct {
	Group  // the group I belong to
	Kind   string
	Spec   Field
	Status *Field

	// The set of additional printer columns to apply to the CustomResourceDefinition
	AdditionalPrinterColumns []apiextv1beta1.CustomResourceColumnDefinition
}

type Field struct {
	Type Type
}

type Type struct {
	// name of the type.
	Name string

	// proto message for the type, if the proto message is compiled with Autopilot
	Message proto.Message

	/*

		The go package containing the type of the file relative path to where the status and spec field can be found, if different than group root.
		Will be set automatically for proto-based types.

		If unset, AutoPilot uses the default types package for the type.
	*/
	GoPackage string
}
