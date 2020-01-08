package render

import (
	"github.com/solo-io/autopilot/codegen/util"
)

// runs kubernetes code-generator.sh
// cannot be used to write output to memory
// also generates deecopy code
func KubeCodegen(apiDir string, group Group) error {
	return util.KubeCodegen(group.Group, group.Version, apiDir)
}
