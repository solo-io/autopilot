package model

import "k8s.io/apimachinery/pkg/runtime/schema"

type Group struct {
	// the group version of the group
	schema.GroupVersion

	// the go module this group belongs to
	Module string

	// the kinds in the group
	Resources []Resource

	// Should we generate kubernetes manifests?
	RenderManifests bool

	// Should we generate kubernetes Go structs?
	RenderTypes bool

	// Should we generate kubernetes Go clients?
	RenderClients bool

	// Should we generate kubernetes Go controllers?
	RenderController bool
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
}

type Field struct {
	Type string
}
