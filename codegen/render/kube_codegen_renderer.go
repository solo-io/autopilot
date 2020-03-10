package render

import (
	"log"

	"github.com/solo-io/autopilot/codegen/model"

	"github.com/solo-io/autopilot/codegen/util"
)

// runs kubernetes code-generator.sh
// cannot be used to write output to memory
// also generates deecopy code
func KubeCodegen(group Group) error {
	log.Printf("Running Kubernetes Codegen for %v", group.Group)
	if group.RenderTypes && len(group.Generators) == 0 {
		group.Generators = append(group.Generators, model.GeneratorType_Deepcopy)
	}
	return util.KubeCodegen(group.Group, group.Version, group.ApiRoot, group.Generators.Strings())
}
