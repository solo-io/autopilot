package processing

import (
	aliases "github.com/solo-io/autopilot/pkg/aliases"
	metrics "github.com/solo-io/autopilot/pkg/metrics"
)

type Inputs struct {
	Metrics metrics.Metrics
}

type Outputs struct {
	Pods aliases.Pods
}
