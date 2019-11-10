package logs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/spf13/cobra"
)

var (
	namespace string
	follow    bool
)

func NewCmd() *cobra.Command {
	logsCmd := &cobra.Command{
		Use:   "logs <image>",
		Short: "Pipe Logs from the Operator to Stderr/Stdout",
		Long: `
`,
		RunE: logsFunc,
	}
	logsCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace to which to logs the operator")
	logsCmd.PersistentFlags().BoolVarP(&follow, "follow", "f", false, "Follow logs (Ctrl+C to stop)")

	return logsCmd
}

func logs(operatorName string) error {
	args := []string{
		"logs",
		"-l",
		"name=" + operatorName,
		"-n", namespace,
	}

	if follow {
		args = append(args, "-f")
	}

	logs := exec.Command("kubectl", args...)
	logs.Stderr = os.Stderr
	logs.Stdout = os.Stdout
	if err := logs.Run(); err != nil {
		return err
	}

	return nil
}

func logsFunc(cmd *cobra.Command, args []string) error {
	util.MustInProjectRoot()

	cfg := codegen.MustLoad()

	if namespace == "" {
		namespace = cfg.OperatorName
	}

	if err := logs(cfg.OperatorName); err != nil {
		return fmt.Errorf("failed to get operator logs: (%v)", err)
	}
	return nil
}
