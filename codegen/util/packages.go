package util

import (
	"github.com/solo-io/autopilot/codegen/model"
	"strings"
)

// gets the go package for the group
func GoPackage(grp model.Group, module, apiRoot string) string {
	apiRoot = strings.Trim(apiRoot, "/")

	return strings.Join([]string{
		module,
		apiRoot,
		grp.Group,
		grp.Version,
	}, "/")
}
