package codegen

import (
	"bytes"
	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/autopilot/api/v1"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/templates"
	"github.com/solo-io/autopilot/codegen/templates/deploy"
	"github.com/solo-io/autopilot/pkg/defaults"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"text/template"

	"github.com/solo-io/autopilot/codegen/util"
	"sigs.k8s.io/yaml"
)

// load the default config or die
func MustLoad() *model.ProjectData {
	data, err := Load(defaults.AutoPilotFile, defaults.OperatorFile)
	if err != nil {
		logrus.Fatalf("failed to load autopilot.yaml: %v", err)
	}
	return data
}

// load the provided config as template data
func Load(autoPilotYaml, operatorYaml string) (*model.ProjectData, error) {
	// load autopilot.yaml
	raw, err := ioutil.ReadFile(autoPilotYaml)
	if err != nil {
		return nil, err
	}
	var project v1.AutoPilotProject
	if err := util.UnmarshalYaml(raw, &project); err != nil {
		return nil, err
	}

	// load autopilot-operator.yaml
	raw, err = ioutil.ReadFile(operatorYaml)
	if err != nil {
		return nil, err
	}
	var operator v1.AutoPilotOperator
	if err := util.UnmarshalYaml(raw, &operator); err != nil {
		return nil, err
	}

	if err := os.Chdir(filepath.Dir(autoPilotYaml)); err != nil {
		return nil, err
	}

	return model.NewTemplateData(project, operator)
}

type GenFile struct {
	OutPath       string
	SkipOverwrite bool
	Permission    os.FileMode

	// set by generate
	Content string

	// either TemplatePath or TemplateFunc is set
	TemplatePath string
	TemplateFunc templates.TemplateFunc
}

func (gf GenFile) GenProjectFile(data *model.ProjectData) (string, error) {
	if gf.TemplatePath != "" {
		return renderProjectFile(data, gf.TemplatePath)
	}
	return gf.genTemplateFunc(data)
}

func (gf GenFile) GenPhaseFile(data *model.ProjectData, phase model.Phase) (string, error) {
	if gf.TemplatePath != "" {
		return renderPhaseFile(data, phase, gf.TemplatePath)
	}
	return gf.genTemplateFunc(data)
}

func (gf GenFile) genTemplateFunc(data *model.ProjectData) (string, error) {
	obj := gf.TemplateFunc(data).(metav1.Object)

	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels["app.kubernetes.io/name"] = data.OperatorName
	labels["app.kubernetes.io/name"] = data.OperatorName

	obj.SetLabels(labels)

	yam, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	var v map[string]interface{}

	if err := yaml.Unmarshal(yam, &v); err != nil {
		return "", err
	}

	delete(v, "status")
	// why do we have to do this? Go problem???
	meta := v["metadata"].(map[string]interface{})

	delete(meta, "creationTimestamp")
	v["metadata"] = meta

	if spec, ok := v["spec"].(map[string]interface{}); ok {
		if template, ok := spec["template"].(map[string]interface{}); ok {
			if meta, ok := template["metadata"].(map[string]interface{}); ok {
				delete(meta, "creationTimestamp")
				template["metadata"] = meta
				spec["template"] = template
				v["spec"] = spec
			}
		}
	}

	yam, err = yaml.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(yam), nil
}

func projectFiles(data *model.ProjectData) []GenFile {
	files := []GenFile{
		// code
		{OutPath: filepath.Join(data.ProjectPackage, "cmd/"+data.OperatorName+"/main.go"), TemplatePath: "code/main.gotmpl"},
		{OutPath: filepath.Join(data.SchedulerImportPath, "scheduler.go"), TemplatePath: "code/scheduler.gotmpl"},
		{OutPath: filepath.Join(data.ParametersImportPath, "parameters.go"), TemplatePath: "code/parameters.gotmpl"},

		{OutPath: filepath.Join(data.TypesImportPath, "doc.go"), TemplatePath: "code/doc.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "phases.go"), TemplatePath: "code/phases.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "register.go"), TemplatePath: "code/register.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "spec.go"), TemplatePath: "code/spec.gotmpl", SkipOverwrite: true},
		{OutPath: filepath.Join(data.TypesImportPath, "types.go"), TemplatePath: "code/types.gotmpl"},

		// build
		{OutPath: filepath.Join(data.ProjectPackage, "build", "Dockerfile"), TemplatePath: "build/Dockerfile.tmpl"},
		{OutPath: filepath.Join(data.ProjectPackage, "build", "bin", "user_setup"), TemplatePath: "build/user_setup.tmpl", Permission: 0777},
		{OutPath: filepath.Join(data.ProjectPackage, "build", "bin", "entrypoint"), TemplatePath: "build/entrypoint.tmpl", Permission: 0777},

		// deploy
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "crd.yaml"), TemplateFunc: deploy.CustomResourceDefinition},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "deployment-namespace-scoped.yaml"), TemplateFunc: deploy.NamespaceScopedDeployment},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "deployment-cluster-scoped.yaml"), TemplateFunc: deploy.ClusterScopedDeployment},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "configmap.yaml"), TemplateFunc: deploy.ConfigMap},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "role.yaml"), TemplateFunc: deploy.Role},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "rolebinding.yaml"), TemplateFunc: deploy.RoleBinding},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "clusterrole.yaml"), TemplateFunc: deploy.ClusterRole},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "clusterrolebinding.yaml"), TemplateFunc: deploy.ClusterRoleBinding},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "service_account.yaml"), TemplateFunc: deploy.ServiceAccount},

		// hack
		{OutPath: filepath.Join(data.ProjectPackage, "hack/create_cr_yaml.go"), TemplatePath: "hack/create_cr_yaml.gotmpl", SkipOverwrite: true},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", data.KindLower+"_example.yaml"), TemplateFunc: deploy.CustomResource},
	}

	if data.EnableFinalizer {
		files = append(files, GenFile{
			OutPath: filepath.Join(data.FinalizerImportPath, "finalizer.go"), TemplatePath: "code/finalizer.gotmpl", SkipOverwrite: true,
		})
	}

	if data.NeedsPrometheus() {
		files = append(files, GenFile{
			OutPath: filepath.Join(data.ProjectPackage, "deploy", "prometheus.yaml"), TemplatePath: "deploy/prometheus.yamltmpl",
		})
	}

	return files
}

func phaseFiles(data *model.ProjectData, phase model.Phase) []GenFile {
	return []GenFile{
		{OutPath: filepath.Join(data.ProjectPackage, "pkg", "workers", model.WorkerImportPrefix(phase), "inputs_outputs.go"), TemplatePath: "code/inputs_outputs.gotmpl"},
		{OutPath: filepath.Join(data.ProjectPackage, "pkg", "workers", model.WorkerImportPrefix(phase), "worker.go"), TemplatePath: "code/worker.gotmpl", SkipOverwrite: true},
	}
}

func Generate(data *model.ProjectData) ([]GenFile, error) {
	var files []GenFile
	for _, projectFile := range projectFiles(data) {
		contents, err := projectFile.GenProjectFile(data)
		if err != nil {
			return nil, err
		}

		projectFile.Content = contents
		files = append(files, projectFile)
	}

	for _, phase := range data.AutoPilotProject.Phases {
		phase := model.MustPhase(data, phase)
		for _, phaseFile := range phaseFiles(data, phase) {
			contents, err := phaseFile.GenPhaseFile(data, phase)
			if err != nil {
				return nil, err
			}
			phaseFile.Content = contents
			files = append(files, phaseFile)
		}
	}

	return files, nil
}

func renderProjectFile(data *model.ProjectData, templateFile string) (string, error) {
	fullPath := filepath.Join(autopilotRoot(), "codegen", "templates", templateFile)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templateFile).Funcs(data.Funcs()).Parse(string(content))
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Funcs(data.Funcs()).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderPhaseFile(data *model.ProjectData, phase model.Phase, templateFile string) (string, error) {
	fullPath := filepath.Join(autopilotRoot(), "codegen", "templates", templateFile)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templateFile).Funcs(data.Funcs()).Parse(string(content))
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Funcs(data.Funcs()).Execute(buf, phase); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func autopilotRoot() string {
	return filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "solo-io", "autopilot")
}
