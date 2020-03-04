package render

import (
	"log"

	"github.com/solo-io/autopilot/pkg/utils"

	"github.com/solo-io/autopilot/codegen/util"
)

// runs kubernetes code-generator.sh
// cannot be used to write output to memory
// also generates deecopy code
func KubeCodegen(group Group) error {
	log.Printf("Running Kubernetes Codegen for %v", group.Group)
	if group.RenderTypes && !utils.ContainsString(group.Generators, "deepcopy") {
		group.Generators = append(group.Generators, "deepcopy")
	}
	return util.KubeCodegen(group.Group, group.Version, group.ApiRoot, group.Generators)
}
