// This package defines defaults which are built-into the system
package defaults

var (
	// configuration file for the autopilot CLI
	// this file will be used to generate and re-generate the autopilot operator
	// it is expected to live at the root of the project repository
	AutoPilotFile  = "autopilot.yaml"

	// configuration file for the autopilot operator
	// these files will be loaded at boot time by the operator
	ConfigFile = "autopilot-operator.yaml"
)
