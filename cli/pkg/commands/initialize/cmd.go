package initialize

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/pkg/config"
	"github.com/solo-io/autopilot/pkg/defaults"
	"github.com/spf13/cobra"
)

var (
	group   string
	version string
)

func NewCmd() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Initialize a new project",
		Long: `The autopilot init command creates a new project for the given name.
`,
		RunE: initFunc,
	}
	genCmd.PersistentFlags().StringVar(&group, "group", "example.io", "API Group for the Top-Level CRD")
	genCmd.PersistentFlags().StringVar(&version, "version", "v1", "API Version for the Top-Level CRD")

	return genCmd
}

func initFunc(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("command %s requires exactly one argument", cmd.CommandPath())
	}

	return initAutopilotProject(args[0])
}

func initAutopilotProject(name string) error {
	kind := strcase.ToCamel(name)
	lowerName := strings.ToLower(name)
	if err := os.MkdirAll(lowerName, 0777); err != nil {
		return err
	}

	cfg := &v1.AutoPilotProject{
		OperatorName: lowerName + "-operator",
		ApiVersion:   group + "/" + version,
		Kind:         kind,
		Phases: []*v1.Phase{
			{
				Name:        "Initializing",
				Description: kind + " has begun initializing",
				Outputs:     []string{model.TrafficSplits.LowerName},
				Initial:     true,
			},
			{
				Name:        "Processing",
				Description: kind + " has begun processing",
				Inputs:      []string{model.Metrics.LowerName},
				Outputs:     []string{model.TrafficSplits.LowerName},
			},
			{
				Name:        "Finished",
				Description: kind + " has finished",
				Final:       true,
			},
		},
	}
	yam, err := util.MarshalYaml(cfg)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(lowerName, defaults.AutoPilotFile), yam, 0644); err != nil {
		return err
	}

	operator := &config.DefaultConfig
	yam, err = util.MarshalYaml(operator)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(lowerName, defaults.OperatorFile), yam, 0644); err != nil {
		return err
	}

	logrus.Printf("Created new project %v", lowerName)

	return nil
}
