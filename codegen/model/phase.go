package model

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/autopilot/api/v1"
)

// this is the internal representation of the Phases written by the user
// the phase replaces the inputs and output strings with the actual Parameter types
// the phase also contains a reference to the parent project
type Phase struct {
	v1.Phase

	// internal representation of inputs/outputs
	Inputs  []Parameter
	Outputs []Parameter

	// set by load
	Project *ProjectData `json:"-"`
}

// return a model.Phase from a v1.Phase or die
func ModelPhase(data *ProjectData, phase *v1.Phase) (Phase, error) {
	inputs, err := paramsFromNames(phase.Inputs)
	if err != nil {
		return Phase{}, errors.Wrapf(err, "phase %v inputs", phase.Name)
	}
	outputs, err := paramsFromNames(phase.Outputs)
	if err != nil {
		return Phase{}, errors.Wrapf(err, "phase %v outputs", phase.Name)
	}
	return Phase{
		Phase:   *phase,
		Project: data,
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}

func MustPhase(data *ProjectData, phase *v1.Phase) Phase {
	p, err := ModelPhase(data, phase)
	if err != nil {
		logrus.Fatalf("failed to parse phase: %v", err)
	}
	return p
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
