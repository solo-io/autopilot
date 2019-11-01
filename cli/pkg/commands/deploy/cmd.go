package deploy

import (
	"bytes"
	"fmt"
	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	namespace string
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

func deploy(image, namespace string) error {

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

	Kubectl(nil, "create", "ns", namespace)

	for _, man := range manifests {
		log.Printf("Deploying %v", man)
		raw, err := ioutil.ReadFile(filepath.Join("deploy", man))
		if err != nil {
			return err
		}
		if err := KubectlApply(raw, "-n", namespace); err != nil {
			return err
		}
	}

	deploymentBase := "deploy/deployment.yaml"
	log.Printf("Deploying %v", deploymentBase)
	raw, err := deploymentWithImage(deploymentBase, image)
	if err != nil {
		return err
	}
	if err := KubectlApply(raw, "-n", namespace); err != nil {
		return err
	}

	return nil
}

func deployFunc(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("command %s requires exactly one argument", cmd.CommandPath())
	}

	util.MustInProjectRoot()

	if namespace == "" {
		cfg, err := codegen.Load("autopilot.yaml")
		if err != nil {
			return err
		}
		namespace = cfg.OperatorName
	}

	image := args[0]

	log.Infof("Deploying Operator with image %s", image)

	if err := deploy(image, namespace); err != nil {
		return fmt.Errorf("failed to deploy operator with image %s: (%v)", image, err)
	}

	log.Info("Operator deployment complete.")
	return nil
}

func KubectlApply(manifest []byte, extraArgs ...string) error {
	return Kubectl(bytes.NewBuffer(manifest), append([]string{"apply", "-f", "-"}, extraArgs...)...)
}

func Kubectl(stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	return kubectl.Run()
}
