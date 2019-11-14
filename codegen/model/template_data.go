package model

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/util"
)

// ProjectData is used for rendering templates and generating files
// It is loaded from the user config,
type ProjectData struct {
	v1.AutopilotProject
	v1.AutopilotOperator

	// packr box in which text templates are stored
	// read from codegen/templates
	Templates packr.Box

	// internal implementation of phases
	Phases []Phase `json:"phases"`

	ProjectPackage string // e.g. "github.com/solo-io/autopilot/examples/promoter"

	Group   string // e.g. "mesh.demos.io"
	Version string // e.g. "v1"

	// TODO: refactor these duplicates struct
	// will require refactoring templates
	TypesRelativePath      string // e.g. "pkg/apis/canaries/v1"
	SchedulerRelativePath  string // e.g. "pkg/scheduler"
	FinalizerRelativePath  string // e.g. "pkg/finalizer"
	ParametersRelativePath string // e.g. "pkg/parameters"
	MetricsRelativePath    string // e.g. "pkg/metrics"

	TypesImportPath      string // e.g. "github.com/yourorg/yourproject/pkg/apis/canaries/v1"
	SchedulerImportPath  string // e.g. "github.com/yourorg/yourproject/pkg/scheduler"
	FinalizerImportPath  string // e.g. "github.com/yourorg/yourproject/pkg/finalizer"
	ParametersImportPath string // e.g. "github.com/yourorg/yourproject/pkg/parameters"
	MetricsImportPath    string // e.g. "github.com/yourorg/yourproject/pkg/metrics"

	KindLowerCamel  string // e.g. "YourKind"
	KindLower       string // e.g. "yourresource"
	KindLowerPlural string // e.g. "yourresources"
}

func NewTemplateData(project v1.AutopilotProject, operator v1.AutopilotOperator, templates packr.Box) (*ProjectData, error) {
	projectGoPkg := util.GetGoPkg()

	for _, q := range DefaultQueries {
		q := q // Go!!
		project.Queries = append(project.Queries, &q)
	}

	apiVersionParts := strings.Split(project.ApiVersion, "/")

	if len(apiVersionParts) != 2 {
		return nil, fmt.Errorf("%v must be format groupname/version", apiVersionParts)
	}

	apiGroup := apiVersionParts[0]
	version := apiVersionParts[1]

	// register all custom types
	for _, custom := range project.CustomParameters {
		Register(*custom)
	}

	data := &ProjectData{
		AutopilotProject:     project,
		AutopilotOperator:    operator,
		Templates:            templates,
		ProjectPackage:       projectGoPkg,
		Group:                apiGroup,
		Version:              version,
		TypesImportPath:      filepath.Join(projectGoPkg, TypesRelativePath(project.Kind, version)),
		SchedulerImportPath:  filepath.Join(projectGoPkg, SchedulerRelativePath),
		FinalizerImportPath:  filepath.Join(projectGoPkg, FinalizerRelativePath),
		ParametersImportPath: filepath.Join(projectGoPkg, ParametersRelativePath),
		MetricsImportPath:    filepath.Join(projectGoPkg, MetricsRelativePath),
		KindLowerCamel:       strcase.ToLowerCamel(project.Kind),
		KindLower:            strings.ToLower(project.Kind),
		KindLowerPlural:      pluralize.NewClient().Plural(strings.ToLower(project.Kind)),
	}

	// required for use by worker template
	for _, phase := range project.Phases {
		inputs, err := paramsFromNames(phase.Inputs)
		if err != nil {
			return nil, errors.Wrapf(err, "phase %v inputs", phase.Name)
		}
		outputs, err := paramsFromNames(phase.Outputs)
		if err != nil {
			return nil, errors.Wrapf(err, "phase %v outputs", phase.Name)
		}
		data.Phases = append(data.Phases, Phase{
			Phase:   *phase,
			Project: data,
			Inputs:  inputs,
			Outputs: outputs,
		})
	}

	return data, nil
}

var invalidMetricsOutputParamErr = fmt.Errorf("metrics is not a valid output parameter")

func (d *ProjectData) Validate() error {
	for _, phase := range d.Phases {
		for _, out := range phase.Outputs {
			if out.Equals(Metrics) {
				return invalidMetricsOutputParamErr
			}
		}
	}
	return nil
}

func (d *ProjectData) NeedsMetrics() bool {
	for _, phase := range d.Phases {
		for _, in := range phase.Inputs {
			if in.Equals(Metrics) {
				return true
			}
		}
	}
	return false
}

// operator-local prometheus is currently disabled.
// use prometheus.istio-system instead.
func (d *ProjectData) NeedsPrometheus() bool {
	return false
}

func (d *ProjectData) UniqueOutputs() []Parameter {
	var unique []Parameter
	addParam := func(param Parameter) {
		for _, p := range unique {
			if p.Equals(param) {
				return
			}
		}
		unique = append(unique, param)
	}
	for _, phase := range d.Phases {
		for _, out := range phase.Outputs {
			addParam(out)
		}
	}
	return unique
}

func (d *ProjectData) UniqueParams() []Parameter {
	var unique []Parameter
	addParam := func(param Parameter) {
		for _, p := range unique {
			if p.Equals(param) {
				return
			}
		}
		unique = append(unique, param)
	}
	for _, phase := range d.Phases {
		for _, p := range phase.Inputs {
			addParam(p)
		}
		for _, p := range phase.Outputs {
			addParam(p)
		}
	}
	return unique
}
