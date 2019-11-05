package model

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type Project struct {
	OperatorName string  `json:"operatorName"`
	ApiVersion   string  `json:"apiVersion"`
	Kind         string  `json:"kind"`
	Phases       []Phase `json:"phases"`

	// enable use of a Finalizer to handle object deletion
	EnableFinalizer bool `json:"enableFinalizer"`
}

type Phase struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Initial     bool        `json:"initial,omitempty"`
	Inputs      []Parameter `json:"inputs,omitempty"`
	Outputs     []Parameter `json:"outputs,omitempty"`

	// set by load
	Project *TemplateData `json:"-"`
}

// with parameter as a string
type userPhase struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Initial     bool     `json:"initial,omitempty"`
	Inputs      []string `json:"inputs,omitempty"`
	Outputs     []string `json:"outputs,omitempty"`
}

func paramNames(params []Parameter) []string {
	var names []string
	for _, p := range params {
		names = append(names, p.LowerName)
	}
	return names
}

func paramsFromNames(names []string) ([]Parameter, error) {
	getParam := func(name string) *Parameter {
		for _, p := range Parameters {
			if p.LowerName == name {
				return &p
			}
		}
		return nil
	}

	var params []Parameter
	for _, n := range names {
		p := getParam(n)
		if p == nil {
			return nil, errors.Errorf("no parameter named %v", n)
		}
		params = append(params, *p)
	}
	return params, nil
}

func (p *Phase) MarshalJSON() ([]byte, error) {
	user := &userPhase{
		Name:        p.Name,
		Description: p.Description,
		Initial:     p.Initial,
		Inputs:      paramNames(p.Inputs),
		Outputs:     paramNames(p.Outputs),
	}
	return json.Marshal(user)
}

func (p *Phase) UnmarshalJSON(b []byte) error {
	var user userPhase
	if err := json.Unmarshal(b, &user); err != nil {
		return err
	}
	inputs, err := paramsFromNames(user.Inputs)
	if err != nil {
		return err
	}
	outputs, err := paramsFromNames(user.Outputs)
	if err != nil {
		return err
	}

	*p = Phase{
		Name:        user.Name,
		Description: user.Description,
		Initial:     user.Initial,
		Inputs:      inputs,
		Outputs:     outputs,
	}
	return nil
}

type Parameter struct {
	LowerName    string
	SingleName   string
	PluralName   string
	ImportPrefix string
	Package      string
	ApiGroup     string
}

func (p Parameter) String() string {
	return string(p.LowerName)
}

// registered parameters
var Parameters []Parameter

func register(param Parameter) Parameter {
	Parameters = append(Parameters, param)
	return param
}

var (
	ReplicaSets = register(Parameter{
		LowerName:    "replicasets",
		PluralName:   "ReplicaSets",
		SingleName:   "ReplicaSet",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "extensions",
	})
	ConfigMaps = register(Parameter{
		LowerName:    "configmaps",
		PluralName:   "ConfigMaps",
		SingleName:   "ConfigMap",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "",
	})
	Pods = register(Parameter{
		LowerName:    "pods",
		PluralName:   "Pods",
		SingleName:   "Pod",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "",
	})
	Deployments = register(Parameter{
		LowerName:    "deployments",
		PluralName:   "Deployments",
		SingleName:   "Deployment",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "apps",
	})
	Services = register(Parameter{
		LowerName:    "services",
		PluralName:   "Services",
		SingleName:   "Service",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "",
	})
	TrafficSplits = register(Parameter{
		LowerName:    "trafficsplits",
		PluralName:   "TrafficSplits",
		SingleName:   "TrafficSplit",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "split.smi-spec.io",
	})
	VirtualServices = register(Parameter{
		LowerName:    "virtualservices",
		PluralName:   "VirtualServices",
		SingleName:   "VirtualService",
		ImportPrefix: "aliases",
		Package:      "github.com/solo-io/autopilot/pkg/aliases",
		ApiGroup:     "networking.istio.io",
	})
	Metrics = register(Parameter{
		LowerName:    "metrics",
		PluralName:   "Metrics",
		SingleName:   "Metric",
		ImportPrefix: "metrics",
		Package:      "github.com/solo-io/autopilot/pkg/metrics",
	})
)
