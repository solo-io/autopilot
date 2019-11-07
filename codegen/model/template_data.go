package model

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/util"
	"path/filepath"
	"strings"
)

// used for rendering templates
type ProjectData struct {
	v1.AutoPilotProject
	v1.AutoPilotOperator

	// internal implementation of phases
	Phases []Phase `json:"phases"`

	ProjectPackage string // e.g. "github.com/solo-io/autopilot/examples/promoter"

	Group   string // e.g. "mesh.demos.io"
	Version string // e.g. "v1"

	TypesImportPath      string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/apis/canaries/v1"
	SchedulerImportPath  string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/scheduler"
	FinalizerImportPath  string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/finalizer"
	ParametersImportPath string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/parameters"

	KindLowerCamel  string // e.g. "canaryResource"
	KindLower       string // e.g. "canaryresource"
	KindLowerPlural string // e.g. "canaryresources"
}

func NewTemplateData(project v1.AutoPilotProject, operator v1.AutoPilotOperator) (*ProjectData, error) {
	projectGoPkg := util.GetGoPkg()

	apiVersionParts := strings.Split(project.ApiVersion, "/")

	if len(apiVersionParts) != 2 {
		return nil, fmt.Errorf("%v must be format groupname/version", apiVersionParts)
	}

	c := pluralize.NewClient()

	apiGroup := apiVersionParts[0]
	apiVersion := apiVersionParts[1]

	apiImportPath := filepath.Join(projectGoPkg, "pkg", "apis", strings.ToLower(c.Plural(project.Kind)), apiVersion)
	schedulerImportPath := filepath.Join(projectGoPkg, "pkg", "scheduler")
	finalizerImportPath := filepath.Join(projectGoPkg, "pkg", "finalizer")
	parametersImportPath := filepath.Join(projectGoPkg, "pkg", "parameters")

	// register all custom types
	for _, custom := range project.CustomParameters {
		Register(*custom)
	}

	data := &ProjectData{
		AutoPilotProject:     project,
		AutoPilotOperator:    operator,
		ProjectPackage:       projectGoPkg,
		Group:                apiGroup,
		Version:              apiVersion,
		TypesImportPath:      apiImportPath,
		SchedulerImportPath:  schedulerImportPath,
		FinalizerImportPath:  finalizerImportPath,
		ParametersImportPath: parametersImportPath,
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
