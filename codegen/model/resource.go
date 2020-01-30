package model

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Group struct {
	// the group version of the group
	schema.GroupVersion

	// the go  modulethis group belongs to
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
	// leave empty if same as output dir
	Package string
}

type Field struct {
	Type string
}
