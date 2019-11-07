package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	v1 "github.com/solo-io/autopilot/api/v1"
)

// this is the internal representation of the Project written by the user
// for convenience, the user object is simplified
// as defined by autopilot.proto.
// conversion is handled by custom Marshal/Unmarshal functions
type Project struct {
	OperatorName string  `json:"operatorName"`
	ApiVersion   string  `json:"apiVersion"`
	Kind         string  `json:"kind"`
	Phases       []Phase `json:"phases"`

	// custom parameters specified by the user
	// can be used as inputs or outputs
	CustomParameters []Parameter `json:"customParameters"`

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
	user := &v1.Phase{
		Name:        p.Name,
		Description: p.Description,
		Initial:     p.Initial,
		Inputs:      paramNames(p.Inputs),
		Outputs:     paramNames(p.Outputs),
	}
	return json.Marshal(user)
}

func (p *Phase) UnmarshalJSON(b []byte) error {
	var user v1.Phase
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
