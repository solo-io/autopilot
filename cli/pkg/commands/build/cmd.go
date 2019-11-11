package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	imageBuildArgs string
	goBuildArgs    string
)

func NewCmd() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build <image>",
		Short: "Compiles code and builds artifacts",
		Long: `The operator-sdk build command compiles the Operator code into an executable binary
and generates the Dockerfile manifest.

<image> is the container image to be built, e.g. "quay.io/example/operator:v0.0.1".
This image will be automatically set in the deployment manifests.

After build completes, the image would be built locally in docker. Then it needs to
be pushed to remote registry.
For example:
	$ operator-sdk build quay.io/example/operator:v0.0.1
	$ docker push quay.io/example/operator:v0.0.1
`,
		RunE: buildFunc,
	}
	buildCmd.Flags().StringVar(&imageBuildArgs, "image-build-args", "", "Extra image build arguments as one string such as \"--build-arg https_proxy=$https_proxy\"")
	buildCmd.Flags().StringVar(&goBuildArgs, "go-build-args", "", "Extra Go build arguments as one string such as \"-ldflags -X=main.xyz=abc\"")
	return buildCmd
}

func createBuildCommand(context, dockerFile, image string, imageBuildArgs ...string) (*exec.Cmd, error) {
	args := []string{"build", "-f", dockerFile, "-t", image}

	for _, bargs := range imageBuildArgs {
		if bargs != "" {
			splitArgs := strings.Fields(bargs)
			args = append(args, splitArgs...)
		}
	}

	args = append(args, context)

	return exec.Command("docker", args...), nil
}

func buildFunc(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("command %s requires exactly one argument", cmd.CommandPath())
	}

	util.MustInProjectRoot()
	goBuildEnv := append(os.Environ(), "GOOS=linux")

	if value, ok := os.LookupEnv("GOARCH"); ok {
		goBuildEnv = append(goBuildEnv, "GOARCH="+value)
	} else {
		goBuildEnv = append(goBuildEnv, "GOARCH=amd64")
	}

	// If CGO_ENABLED is not set, set it to '0'.
	if _, ok := os.LookupEnv("CGO_ENABLED"); !ok {
		goBuildEnv = append(goBuildEnv, "CGO_ENABLED=0")
	}

	absProjectPath := util.MustGetwd()

	goArgs := []string{}

	if goBuildArgs != "" {
		splitArgs := strings.Fields(goBuildArgs)
		goArgs = append(goArgs, splitArgs...)
	}

	data := codegen.MustLoad()

	opts := util.GoCmdOptions{
		BinName:     filepath.Join(absProjectPath, "build", "_output", "bin", data.OperatorName),
		PackagePath: filepath.Join(util.GetGoPkg(), "cmd", data.OperatorName),
		Args:        goArgs,
		Env:         goBuildEnv,
	}
	if err := util.GoBuild(opts); err != nil {
		return fmt.Errorf("failed to build operator binary: (%v)", err)
	}

	image := args[0]

	log.Infof("Building OCI image %s", image)

	buildCmd, err := createBuildCommand(".", "build/Dockerfile", image, imageBuildArgs)
	if err != nil {
		return err
	}

	if err := util.ExecCmd(buildCmd); err != nil {
		return fmt.Errorf("failed to output build image %s: (%v)", image, err)
	}

	log.Info("Operator build complete.")
	return nil
}
