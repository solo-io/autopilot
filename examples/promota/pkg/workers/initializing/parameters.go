package initializing

import (
	aliases "github.com/solo-io/autopilot/pkg/aliases"
)

type Inputs struct {
	Deployments aliases.Deployments
}

type Outputs struct {
	Deployments   aliases.Deployments
	Services      aliases.Services
	TrafficSplits aliases.TrafficSplits
}
