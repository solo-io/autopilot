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
}

type Parameter string

const (
	Deployments   Parameter = "deployments"
	Services      Parameter = "services"
	TrafficSplits Parameter = "trafficsplits"
	Metrics       Parameter = "metrics"
)
