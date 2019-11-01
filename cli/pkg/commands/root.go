package commands

import (
	"github.com/solo-io/autopilot/cli/pkg/commands/deploy"
	"github.com/solo-io/autopilot/cli/pkg/commands/generate"
	"github.com/solo-io/autopilot/cli/pkg/commands/initialize"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/autopilot/cli/pkg/commands/build"
)

func AutoPilotCli() *cobra.Command {
	root := &cobra.Command{
		Use:     "autopilot",
		Aliases: []string{"ap"},
		Short:   "An SDK for building Service Mesh Operators with ease",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if viper.GetBool("verbose") {
				if err := util.SetGoVerbose(); err != nil {
					log.Fatalf("Could not set GOFLAGS: (%v)", err)
				}
				log.SetLevel(log.DebugLevel)
				log.Debug("Debug logging is set")
			}
		},
	}
	root.AddCommand(initialize.NewCmd())
	root.AddCommand(generate.NewCmd())
	root.AddCommand(build.NewCmd())
	root.AddCommand(deploy.NewCmd())

	root.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	if err := viper.BindPFlags(root.PersistentFlags()); err != nil {
		log.Fatalf("Failed to bind root flags: %v", err)
	}

	return root
}
