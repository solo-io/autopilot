package deploy

import (
	"fmt"
	"github.com/solo-io/autopilot/cli/pkg/utils"
	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	namespace  string
	deletePods bool
)

func NewCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy <image>",
		Short: "Deploys the Operator with the provided image to Kubernetes",
		Long: `
`,
		RunE: deployFunc,
	}
	deployCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace to which to deploy the operator")
	deployCmd.PersistentFlags().BoolVarP(&deletePods, "deletepods", "d", false, "Delete existing pods after pushing images (to force Kubernetes to pull the newly pushed image)")
	return deployCmd
}

func deploymentWithImage(deploymentBase, image string) ([]byte, error) {
	raw, err := ioutil.ReadFile(deploymentBase)
	if err != nil {
		return nil, err
	}
	withImage := strings.ReplaceAll(string(raw), "REPLACE_IMAGE", image)
	return []byte(withImage), nil
}

func deploy(operatorName, image, namespace string, deletePods bool) error {

	log.Printf("Pushing image %v", image)
	push := exec.Command("docker", "push", image)
	push.Stderr = os.Stderr
	push.Stdout = os.Stdout
	if err := push.Run(); err != nil {
		return err
	}

	manifests := []string{
		"crd.yaml",
		"role.yaml",
		"rolebinding.yaml",
		"service_account.yaml",
	}

	utils.Kubectl(nil, "create", "ns", namespace)

	for _, man := range manifests {
		log.Printf("Deploying %v", man)
		raw, err := ioutil.ReadFile(filepath.Join("deploy", man))
		if err != nil {
			return err
		}
		if err := utils.KubectlApply(raw, "-n", namespace); err != nil {
			return err
		}
	}

	deploymentBase := "deploy/deployment.yaml"
	log.Printf("Deploying %v", deploymentBase)
	raw, err := deploymentWithImage(deploymentBase, image)
	if err != nil {
		return err
	}
	if err := utils.KubectlApply(raw, "-n", namespace); err != nil {
		return err
	}

	if deletePods {
		if err := utils.Kubectl(nil, "delete", "pod", "-n", namespace, "-l", "name="+operatorName, "--ignore-not-found"); err != nil {
			return err
		}
	}

	return nil
}

func deployFunc(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("command %s requires exactly one argument", cmd.CommandPath())
	}

	util.MustInProjectRoot()

	cfg, err := codegen.Load("autopilot.yaml")
	if err != nil {
		return err
	}

	if namespace == "" {
		namespace = cfg.OperatorName
	}

	image := args[0]

	log.Infof("Deploying Operator with image %s", image)

	if err := deploy(cfg.OperatorName, image, namespace, deletePods); err != nil {
		return fmt.Errorf("failed to deploy operator with image %s: (%v)", image, err)
	}

	log.Info("Operator deployment complete.")
	return nil
}
