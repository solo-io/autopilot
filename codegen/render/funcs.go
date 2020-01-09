package render

import (
	"github.com/iancoleman/strcase"
	"github.com/solo-io/autopilot/codegen/util"
	"strings"
	"text/template"
)

func makeTemplateFuncs(module, apiRoot string) template.FuncMap {
	return template.FuncMap{
		// string utils
		"join":            strings.Join,
		"lower":           strings.ToLower,
		"lower_camel":     strcase.ToLowerCamel,
		"upper_camel":     strcase.ToCamel,
		"snake":           strcase.ToSnake,
		"split":           SplitTrimEmpty,
		"string_contains": strings.Contains,

		// code template funcs
		"group_import_path": func(grp Group) string {
			return util.GoPackage(grp, module, apiRoot)
		},
	}

}
func SplitTrimEmpty(s, sep string) []string {
	return strings.Split(strings.TrimSpace(s), sep)
}
