package util

import (
	"strings"

	"github.com/solo-io/autopilot/codegen/model"
)

// gets the go package for the group
func GoPackage(grp model.Group) string {
	if grp.CustomTypesImportPath != "" {
		return grp.CustomTypesImportPath
	}

	grp.ApiRoot = strings.Trim(grp.ApiRoot, "/")

	s := strings.ReplaceAll(
		strings.Join([]string{
			grp.Module,
			grp.ApiRoot,
			grp.Group,
			grp.Version,
		}, "/"),
		"//", "/",
	)

	return s
}
