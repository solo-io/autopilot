package initialize

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	autopilotversion "github.com/solo-io/autopilot/pkg/version"

	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/pkg/config"
	"github.com/solo-io/autopilot/pkg/defaults"
	"github.com/spf13/cobra"
)

var (
	kind      string
	group     string
	version   string
	module    string
	skipGomod bool
)

func NewCmd() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "init <dir> --kind=<kind> --group=<apigroup> --verison=<apiversion> [--skip-gomod]",
		Short: "Initialize a new project for the given top-level CRD",
		Long: `The autopilot init command creates a project skeleton in the given directory. 
If the directory does not exist, it will be created. 
`,
		RunE: initFunc,
	}
	genCmd.PersistentFlags().StringVar(&kind, "kind", "Example", "Kind (Camel-Cased Name) of Top-Level CRD")
	genCmd.PersistentFlags().StringVar(&group, "group", "examples.io", "API Group for the Top-Level CRD")
	genCmd.PersistentFlags().StringVar(&version, "version", "v1", "API Version for the Top-Level CRD")
	genCmd.PersistentFlags().BoolVarP(&skipGomod, "skip-gomod", "s", false, "skip generating go.mod for project")
	genCmd.PersistentFlags().StringVarP(&module, "module", "m", "", "Sets the name of the module for `go mod init`."+
		"Required if initializing outside your $GOPATH")

	return genCmd
}

func initFunc(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("command %s requires exactly one argument", cmd.CommandPath())
	}

	return initAutopilotProject(args[0])
}

func initAutopilotProject(dir string) error {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	lowerName := strings.ToLower(kind)

	cfg := &v1.AutopilotProject{
		OperatorName: lowerName + "-operator",
		ApiVersion:   group + "/" + version,
		Kind:         kind,
		Phases: []*v1.Phase{
			{
				Name:        "Initializing",
				Description: kind + " has begun initializing",
				Outputs:     []string{model.VirtualServices.LowerName},
				Initial:     true,
			},
			{
				Name:        "Processing",
				Description: kind + " has begun processing",
				Inputs:      []string{model.Metrics.LowerName},
				Outputs:     []string{model.VirtualServices.LowerName},
			},
			{
				Name:        "Finished",
				Description: kind + " has finished",
				Final:       true,
			},
		},
	}

	logrus.Printf("Creating Project Config: %v", cfg)

	yam, err := util.MarshalYaml(cfg)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, defaults.AutopilotFile), yam, 0644); err != nil {
		return err
	}

	operator := &config.DefaultConfig
	yam, err = util.MarshalYaml(operator)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, defaults.OperatorFile), yam, 0644); err != nil {
		return err
	}

	if !skipGomod {
		if err := initGoMod(dir); err != nil {
			return err
		}
		if err := initialGoGet(dir); err != nil {
			return err
		}
	}

	return nil
}

func exists(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}

func initGoMod(dir string) error {
	goMod := filepath.Join(dir, "go.mod")
	if exists(goMod) {
		logrus.Printf("Skipping gomod")
		return nil
	}
	cmd := exec.Command("go", "mod", "init")
	if module != "" {
		cmd.Args = append(cmd.Args, module)
	}
	cmd.Dir = dir
	if err := util.ExecCmd(cmd); err != nil {
		return err
	}
	b, err := ioutil.ReadFile(goMod)
	if err != nil {
		return err
	}
	b = append(b, []byte(goModFooter)...)
	return ioutil.WriteFile(goMod, b, 0644)
}

func initialGoGet(dir string) error {
	versionSuffix := ""
	if autopilotversion.Version != autopilotversion.DevVersion {
		versionSuffix = "@" + autopilotversion.Version
	}
	cmd := exec.Command("go", "get", "-v", "github.com/solo-io/autopilot"+versionSuffix)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Dir = dir
	if err := util.ExecCmd(cmd); err != nil {
		return err
	}
	return nil
}

const goModFooter = `

// Pinned to kubernetes-1.14.1
replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190409022021-00b8e31abe9d
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190510232812-a01b7d5d6c22
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.1
)

replace (
	// Indirect operator-sdk dependencies use git.apache.org, which is frequently
	// down. The github mirror should be used instead.
	// Locking to a specific version (from 'go mod graph'):
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190605231540-b8a4faf68e36
)

// Remove when controller-tools v0.2.2 is released
// Required for the bugfix https://github.com/kubernetes-sigs/controller-tools/pull/322
replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.2-0.20190919011008-6ed4ff330711

require (
	github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf v1.3.2
	github.com/solo-io/autopilot v0.0.0-20191113003254-3821310e6e8c
	go.tmthrgd.dev/gomodpriv v0.0.0-20191024122841-38d17a72f03c // indirect
	istio.io/client-go v0.0.0-20191111192453-21751e6cf0fe
	k8s.io/apimachinery v0.0.0
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/yaml v1.1.0
)
`
