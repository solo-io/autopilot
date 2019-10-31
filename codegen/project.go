package codegen

type Project struct {
	ApiVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	Phases     []Phase `json:"phases"`
}

type Phase struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Inputs      []Parameter `json:"inputs"`
	Outputs     []Parameter `json:"outputs"`
	Initial     bool        `json:"initial"`

	// set by load
	Project *TemplateData
}

type Parameter string

const (
	Deployments   Parameter = "deployments"
	Services      Parameter = "services"
	TrafficSplits Parameter = "trafficsplits"
	Metrics       Parameter = "metrics"
)

type ParameterInfo struct {
	SingleName string
	PluralName string
	ImportPrefix string
	Package string
}

var parameters = map[Parameter]ParameterInfo{
	Deployments: ParameterInfo{
		PluralName:   "Deployments",
		SingleName:   "Deployment",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
	},
	Services: ParameterInfo{
		PluralName:   "Services",
		SingleName:   "Service",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
	},
	TrafficSplits: ParameterInfo{
		PluralName:   "TrafficSplits",
		SingleName:   "TrafficSplit",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
	},
	Metrics: ParameterInfo{
		PluralName: "Metrics",
		SingleName: "Metric",
		ImportPrefix: "metrics",
		Package:      "github.com/solo-io/autopilot/pkg/metrics",
	},
}
