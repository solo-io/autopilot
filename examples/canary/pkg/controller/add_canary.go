package controller

import (
	"github.com/solo-io/autopilot/examples/canary/pkg/controller/canary"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, canary.Add)
}
