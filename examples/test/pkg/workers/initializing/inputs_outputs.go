package initializing

import (
	parameters "github.com/solo-io/autopilot/examples/test/pkg/parameters"
)

type Inputs struct {
	Services parameters.Services
}

type Outputs struct {
	VirtualServices parameters.VirtualServices
	Gateways        parameters.Gateways
}
