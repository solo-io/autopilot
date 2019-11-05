package initializing

import (
	aliases "github.com/solo-io/autopilot/pkg/aliases"
)

type Inputs struct {
	Services aliases.Services
}

type Outputs struct {
	VirtualServices aliases.VirtualServices
}
