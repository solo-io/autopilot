package codegen

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

func (d TemplateData) Funcs() template.FuncMap {
	return template.FuncMap{
		"join":                 strings.Join,
		"lower":                strings.ToLower,
		"lower_camel":          strcase.ToLowerCamel,
		"upper_camel":          strcase.ToCamel,
		"snake":                strcase.ToSnake,
		"output_name":          outputName,
		"has_outputs":          hasOutputs,
		"has_inputs":           hasInputs,
		"is_final":             isFinal,
		"is_metrics":           isMetrics,
		"worker_import_prefix": workerImportPrefix,
		"worker_package":       d.workerPackage,
	}
}

func outputName(param Parameter) string {
	return parameterNames[param]
}

func hasOutputs(phase Phase) bool {
	return len(phase.Outputs) > 0
}

func hasInputs(phase Phase) bool {
	return len(phase.Inputs) > 0
}

func isFinal(phase Phase) bool {
	return !hasOutputs(phase) && !hasInputs(phase)
}

func isMetrics(param Parameter) bool {
	return param == Metrics
}

func workerImportPrefix(phase Phase) string {
	return strings.ToLower(phase.Name)
}

func (d TemplateData) workerPackage(phase Phase) string {
	return filepath.Join(d.ProjectPackage, "pkg", "workers", strings.ToLower(phase.Name))
}
