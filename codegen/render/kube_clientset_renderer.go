package render

import (
	"log"

	"github.com/solo-io/autopilot/codegen/util"
)

// runs kubernetes code-generator.sh
// cannot be used to write output to memory
// also generates deecopy code
func KubeCodegen(group Group) error {
	if !group.RenderClients {
		return nil
	}
	log.Printf("Running Kubernetes Codegen for %v", group.Group)
	return util.KubeCodegen(group.Group, group.Version, group.ApiRoot)
}
