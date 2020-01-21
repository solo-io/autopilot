package render

import (
	"github.com/solo-io/autopilot/codegen/util"
)

// runs kubernetes code-generator.sh
// cannot be used to write output to memory
// also generates deecopy code
func KubeCodegen(group Group) error {
	if !group.RenderClients {
		return nil
	}
	return util.KubeCodegen(group.Group, group.Version, group.ApiRoot)
}
