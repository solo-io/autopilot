package model

import "fmt"

type TemplateData struct {
	Project

	ProjectPackage string // e.g. "github.com/solo-io/autopilot/examples/promoter"

	Group   string // e.g. "mesh.demos.io"
	Version string // e.g. "v1"

	TypesImportPath     string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/apis/canaries/v1"
	SchedulerImportPath string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/scheduler"
	ConfigImportPath    string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/config"
	FinalizerImportPath string // e.g. "github.com/solo-io/autopilot/examples/promoter/pkg/finalizer"
	KindLowerCamel      string // e.g. "canaryResource"
	KindLower           string // e.g. "canaryresource"
	KindLowerPlural     string // e.g. "canaryresources"
}

var invalidMetricsOutputParamErr = fmt.Errorf("metrics is not a valid output parameter")

func (d *TemplateData) Validate() error {
	for _, phase := range d.Phases {
		for _, out := range phase.Outputs {
			if out == Metrics {
				return invalidMetricsOutputParamErr
			}
		}
	}
	return nil
}

func (d *TemplateData) NeedsMetrics() bool {
	for _, phase := range d.Phases {
		for _, in := range phase.Inputs {
			if in == Metrics {
				return true
			}
		}
	}
	return false
}


func (d *TemplateData) UniqueOutputs() []Parameter {
	var unique []Parameter
	addParam := func(param Parameter) {
		for _, p := range unique {
			if p == param {
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
