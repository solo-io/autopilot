package model

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

func (d TemplateData) Funcs() template.FuncMap {
	return template.FuncMap{
		// string utils
		"join":        strings.Join,
		"lower":       strings.ToLower,
		"lower_camel": strcase.ToLowerCamel,
		"upper_camel": strcase.ToCamel,
		"snake":       strcase.ToSnake,
		"split":       splitTrimEmpty,

		// project data
		"param_name_plural":    paramNamePlural,
		"param_name_singular":  paramNameSingular,
		"param_import_prefix":  paramImportPrefix,
		"param_package":        paramPackage,
		"has_outputs":          HasOutputs,
		"has_inputs":           HasInputs,
		"is_final":             isFinal,
		"is_metrics":           isMetrics,
		"worker_import_prefix": WorkerImportPrefix,
		"worker_package":       d.workerPackage,
	}
}

func splitTrimEmpty(s, sep string) []string {
	return strings.Split(strings.TrimSpace(s), sep)
}

func paramNamePlural(param Parameter) string {
	return parameters[param].PluralName
}

func paramNameSingular(param Parameter) string {
	return parameters[param].SingleName
}

func paramImportPrefix(param Parameter) string {
	return parameters[param].ImportPrefix
}

func paramPackage(param Parameter) string {
	return parameters[param].Package
}

func ParamApiVersion(param Parameter) string {
	return parameters[param].ApiVersion
}

func HasOutputs(phase Phase) bool {
	return len(phase.Outputs) > 0
}

func HasInputs(phase Phase) bool {
	return len(phase.Inputs) > 0
}

func isFinal(phase Phase) bool {
	return !HasOutputs(phase) && !HasInputs(phase)
}

func isMetrics(param Parameter) bool {
	return param == Metrics
}

func WorkerImportPrefix(phase Phase) string {
	return strings.ToLower(phase.Name)
}

func (d TemplateData) workerPackage(phase Phase) string {
	return filepath.Join(phase.Project.ProjectPackage, "pkg", "workers", strings.ToLower(phase.Name))
}
