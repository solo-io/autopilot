package test

import (
	"os"
	"testing"

	"github.com/solo-io/autopilot/cli/pkg/commands"
)

func TestIcles(t *testing.T) {
	if err := os.Chdir("/Users/ilackarms/go/src/github.com/solo-io/autopilot/examples/promoter"); err != nil {
		t.Fatal(err)
	}
	if err := AutoPilot("build", "docker.io/ilackarms/aptest"); err != nil {
		t.Fatal(err)
	}
}

func AutoPilot(args ...string) error {
	root := commands.AutoPilotCli()
	root.SetArgs(args)
	return root.Execute()
}
