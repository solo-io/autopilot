package builder

import (
	"os"
	"os/exec"

	"github.com/solo-io/autopilot/codegen/util"
)

// Builder is responsible for building go binaries and docker images
type Builder interface {
	// runs the docker command
	Docker(args ...string) error

	// runs Go Build with the given options
	GoBuild(options util.GoCmdOptions) error
}

// realBuilder executes real docker and go build commands
type realBuilder struct{}

func NewBuilder() *realBuilder {
	return &realBuilder{}
}

func (r *realBuilder) Docker(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()
	return cmd.Run()
}

func (r *realBuilder) GoBuild(options util.GoCmdOptions) error {
	return util.GoBuild(options)
}
