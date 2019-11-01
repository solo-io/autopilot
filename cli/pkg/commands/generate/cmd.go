package generate

import (
	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/spf13/cobra"
)

var (
	forceOverwrite bool
)

func NewCmd() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates code, build, and deployment files",
		Long: `The autopilot generate command (re-)generates the Operator code.

Re-run this command when you have updated your autopilot.yaml or your api's spec.go
`,
		RunE: genFunc,
	}
	genCmd.PersistentFlags().BoolVarP(&forceOverwrite, "overwrite", "f", false, "force overwriting files that are meant to be modified by the user (spec.go, worker.go, etc.)")
	return genCmd
}

func genFunc(cmd *cobra.Command, args []string) error {
	util.MustInProjectRoot()

	return codegen.Run(".", forceOverwrite)
}
