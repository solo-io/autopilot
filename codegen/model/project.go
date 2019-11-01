package model

type Project struct {
	OperatorName string  `json:"operatorName"`
	ApiVersion   string  `json:"apiVersion"`
	Kind         string  `json:"kind"`
	Phases       []Phase `json:"phases"`
}

type Phase struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Inputs      []Parameter `json:"inputs,omitempty"`
	Outputs     []Parameter `json:"outputs,omitempty"`
	Initial     bool        `json:"initial,omitempty"`

	// set by load
	Project *TemplateData `json:"-"`
}

type Parameter string

func (p Parameter) String() string {
	return string(p)
}

const (
	ConfigMaps    Parameter = "configmaps"
	Pods          Parameter = "pods"
	Deployments   Parameter = "deployments"
	Services      Parameter = "services"
	TrafficSplits Parameter = "trafficsplits"
	Metrics       Parameter = "metrics"
)

type ParameterInfo struct {
	SingleName   string
	PluralName   string
	ImportPrefix string
	Package      string
	ApiVersion   string // kube apiversion
}

var parameters = map[Parameter]ParameterInfo{
	ConfigMaps: ParameterInfo{
		PluralName:   "ConfigMaps",
		SingleName:   "ConfigMap",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiVersion:   "",
	},
	Pods: ParameterInfo{
		PluralName:   "Pods",
		SingleName:   "Pod",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiVersion:   "",
	},
	Deployments: ParameterInfo{
		PluralName:   "Deployments",
		SingleName:   "Deployment",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiVersion:   "apps",
	},
	Services: ParameterInfo{
		PluralName:   "Services",
		SingleName:   "Service",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiVersion:   "",
	},
	TrafficSplits: ParameterInfo{
		PluralName:   "TrafficSplits",
		SingleName:   "TrafficSplit",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiVersion:   "split.smi-spec.io",
	},
	Metrics: ParameterInfo{
		PluralName:   "Metrics",
		SingleName:   "Metric",
		ImportPrefix: "metrics",
		Package:      "github.com/solo-io/autopilot/pkg/metrics",
	},
}
