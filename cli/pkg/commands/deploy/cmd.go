package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solo-io/autopilot/cli/pkg/utils"
	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	namespace     string
	image         string
	deletePods    bool
	clusterScoped bool
	push          bool
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
	deployCmd.PersistentFlags().BoolVarP(&clusterScoped, "cluster-scoped", "c", true, "Deploy the operator as a cluster-wide operator. This is required to provide the operator with the ClusterRole required to read and write to other namespaces")
	deployCmd.PersistentFlags().BoolVarP(&push, "push", "p", false, "Push the operator image before deploying. Use in place of `docker push <image>`.")
	return deployCmd
}

func replaceImage(raw []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return readAndReplace(raw, "REPLACE_IMAGE", image)
}

func replaceNamespace(raw []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return readAndReplace(raw, "REPLACE_NAMESPACE", namespace)
}

func readAndReplace(raw []byte, old, new string) ([]byte, error) {
	replaced := strings.ReplaceAll(string(raw), old, new)
	return []byte(replaced), nil
}

func readAndReplaceManifest(file string) ([]byte, error) {
	return replaceNamespace(replaceImage(ioutil.ReadFile(file)))
}

func getManifestsToApply(needsPrometheus bool) []string {
	manifestsToApply := []string{
		"crd.yaml",
		"configmap.yaml",
		"service_account.yaml",
	}
	if clusterScoped {
		manifestsToApply = append(manifestsToApply,
			"clusterrole.yaml",
			"clusterrolebinding.yaml",
			"deployment-all-namespaces.yaml",
		)
	} else {
		manifestsToApply = append(manifestsToApply,
			"role.yaml",
			"rolebinding.yaml",
			"deployment-single-namespace.yaml",
		)
	}

	if needsPrometheus {
		manifestsToApply = append(manifestsToApply,
			"prometheus.yaml",
		)
	}

	return manifestsToApply
}

func deploy(operatorName string, needsPrometheus bool) error {

	if push {
		log.Printf("Pushing image %v", image)
		push := exec.Command("docker", "push", image)
		push.Stderr = os.Stderr
		push.Stdout = os.Stdout
		if err := push.Run(); err != nil {
			return err
		}
	}

	utils.Kubectl(nil, "create", "ns", namespace)

	if deletePods {
		if err := utils.Kubectl(nil, "delete", "pod", "-n", namespace, "-l", "name="+operatorName, "--ignore-not-found"); err != nil {
			return err
		}
	}

	for _, man := range getManifestsToApply(needsPrometheus) {
		log.Printf("Deploying %v", man)

		raw, err := readAndReplaceManifest(filepath.Join("deploy", man))
		if err != nil {
			return err
		}
		if err := utils.KubectlApply(raw, "-n", namespace); err != nil {
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

	cfg := codegen.MustLoad()

	if namespace == "" {
		namespace = cfg.OperatorName
	}

	image = args[0]

	log.Infof("Deploying Operator with image %s", image)

	if err := deploy(cfg.OperatorName, cfg.NeedsPrometheus()); err != nil {
		return fmt.Errorf("failed to deploy operator with image %s: (%v)", image, err)
	}

	log.Info("Operator deployment complete.")
	return nil
}
