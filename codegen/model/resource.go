package model

import "k8s.io/apimachinery/pkg/runtime/schema"

type Group struct {
	schema.GroupVersion
	Resources []Resource
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
