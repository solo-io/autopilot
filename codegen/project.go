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

var parameterNames = map[Parameter]string{
	Deployments:   "Deployments",
	Services:      "Services",
	TrafficSplits: "TrafficSplits",
	Metrics:       "Metrics",
}

var parameterImportPrefixes = map[Parameter]string{
	Deployments:   "aliases",
	Services:      "aliases",
	TrafficSplits: "aliases",
	Metrics:       "metrics",
}

var parameterPackages = map[Parameter]string{
	Deployments:   "github.com/solo-io/autopilot/pkg/aliases",
	Services:      "github.com/solo-io/autopilot/pkg/aliases",
	TrafficSplits: "github.com/solo-io/autopilot/pkg/aliases",
	Metrics:       "github.com/solo-io/autopilot/pkg/metrics",
}
