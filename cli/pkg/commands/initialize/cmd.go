package initialize

import (
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

var (
	forceOverwrite bool
)

func NewCmd() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Initialize a new project",
		Long: `The autopilot init command creates a new project for the given name.
`,
		RunE: initFunc,
	}
	genCmd.PersistentFlags().BoolVarP(&forceOverwrite, "overwrite", "f", false, "force overwriting files that are meant to be modified by the user (spec.go, worker.go, etc.)")
	return genCmd
}

func initFunc(cmd *cobra.Command, args []string) error {

	util.MustInProjectRoot()

	return codegen.Run(".", forceOverwrite)
}

func initAutopilotProject(name string) error {
	kind := strcase.ToCamel(name)
	lowerName := strings.ToLower(name)
	cfg := model.Project{
		OperatorName: lowerName + "-operator",
		ApiVersion:   "autopiot.example.io/v1",
		Kind:         kind,
		Phases: []model.Phase{
			{
				Name:        "Initializing",
				Description: kind + " has begun initializing",
				Outputs:     []model.Parameter{model.TrafficSplits},
				Initial:     true,
			},
			{
				Name:        "Processing",
				Description: kind + " has begun processing",
				Inputs:      []model.Parameter{model.Metrics},
				Outputs:     []model.Parameter{model.TrafficSplits},
			},
			{
				Name:        "Finished",
				Description: kind + " has finished",
			},
		},
	}
	yam, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(lowerName, 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(lowerName, "autopilot.yaml"), yam, 0644); err != nil {
		return err
	}

	logrus.Printf("Created new project %v", lowerName)

	return nil
}
