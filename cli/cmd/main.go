package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/solo-io/autopilot/cli/pkg/commands"
)

func main() {
	root := commands.AutoPilotCli()

	if err := root.Execute(); err != nil {
		log.Fatalf("fatal err: %v", err)
	}
}
