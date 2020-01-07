package templates

import (
	"github.com/iancoleman/strcase"
	"github.com/solo-io/autopilot/codegen/model"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"text/template"
)

type TemplateFunc func(data *model.ProjectData) runtime.Object

var Funcs = template.FuncMap{
	// string utils
	"join":            strings.Join,
	"lower":           strings.ToLower,
	"lower_camel":     strcase.ToLowerCamel,
	"upper_camel":     strcase.ToCamel,
	"snake":           strcase.ToSnake,
	"split":           SplitTrimEmpty,
	"string_contains": strings.Contains,
}

func SplitTrimEmpty(s, sep string) []string {
	return strings.Split(strings.TrimSpace(s), sep)
}
