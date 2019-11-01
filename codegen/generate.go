package codegen

import (
	"bytes"
	"fmt"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/templates"
	"github.com/solo-io/autopilot/codegen/templates/deploy"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/solo-io/autopilot/codegen/util"
	"sigs.k8s.io/yaml"
)

func Load(file string) (*model.TemplateData, error) {
	projData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var project model.Project
	if err := yaml.Unmarshal(projData, &project); err != nil {
		return nil, err
	}

	if err := os.Chdir(filepath.Dir(file)); err != nil {
		return nil, err
	}

	projectGoPkg := util.GetGoPkg()

	apiVersionParts := strings.Split(project.ApiVersion, "/")

	if len(apiVersionParts) != 2 {
		return nil, fmt.Errorf("%v must be format groupname/version", apiVersionParts)
	}

	c := pluralize.NewClient()

	apiGroup := apiVersionParts[0]
	apiVersion := apiVersionParts[1]

	apiImportPath := filepath.Join(projectGoPkg, "pkg", "apis", strings.ToLower(c.Plural(project.Kind)), apiVersion)
	schedulerImportPath := filepath.Join(projectGoPkg, "pkg", "scheduler")
	configImportPath := filepath.Join(projectGoPkg, "pkg", "config")

	data := &model.TemplateData{
		Project:             project,
		ProjectPackage:      projectGoPkg,
		Group:               apiGroup,
		Version:             apiVersion,
		TypesImportPath:     apiImportPath,
		SchedulerImportPath: schedulerImportPath,
		ConfigImportPath:    configImportPath,
		KindLowerCamel:      strcase.ToLowerCamel(project.Kind),
		KindLower:           strings.ToLower(project.Kind),
		KindLowerPlural:     pluralize.NewClient().Plural(strings.ToLower(project.Kind)),
	}

	// required for use by worker template
	for i, phase := range project.Phases {
		phase.Project = data
		project.Phases[i] = phase
	}

	return data, nil
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

func (gf GenFile) GenProjectFile(data *model.TemplateData) (string, error) {
	if gf.TemplatePath != "" {
		return renderProjectFile(data, gf.TemplatePath)
	}
	return gf.genTemplateFunc(data)
}

func (gf GenFile) GenPhaseFile(data *model.TemplateData, phase model.Phase) (string, error) {
	if gf.TemplatePath != "" {
		return renderPhaseFile(data, phase, gf.TemplatePath)
	}
	return gf.genTemplateFunc(data)
}

func (gf GenFile) genTemplateFunc(data *model.TemplateData) (string, error) {
	obj := gf.TemplateFunc(data)

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

func projectFiles(data *model.TemplateData) []GenFile {
	return []GenFile{
		{OutPath: filepath.Join(data.ProjectPackage, "main.go"), TemplatePath: "code/main.gotmpl"},
		{OutPath: filepath.Join(data.SchedulerImportPath, "scheduler.go"), TemplatePath: "code/scheduler.gotmpl"},
		{OutPath: filepath.Join(data.ConfigImportPath, "config.go"), TemplatePath: "code/config.gotmpl", SkipOverwrite: true},
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
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "deployment.yaml"), TemplateFunc: deploy.Deployment},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "role.yaml"), TemplateFunc: deploy.Role},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "rolebinding.yaml"), TemplateFunc: deploy.RoleBinding},
		{OutPath: filepath.Join(data.ProjectPackage, "deploy", "service_account.yaml"), TemplateFunc: deploy.ServiceAccount},
	}
}

func phaseFiles(data *model.TemplateData, phase model.Phase) []GenFile {
	return []GenFile{
		{OutPath: filepath.Join(data.ProjectPackage, "pkg", "workers", model.WorkerImportPrefix(phase), "parameters.go"), TemplatePath: "code/parameters.gotmpl"},
		{OutPath: filepath.Join(data.ProjectPackage, "pkg", "workers", model.WorkerImportPrefix(phase), "worker.go"), TemplatePath: "code/worker.gotmpl", SkipOverwrite: true},
	}
}

func Generate(data *model.TemplateData) ([]GenFile, error) {
	var files []GenFile
	for _, projectFile := range projectFiles(data) {
		contents, err := projectFile.GenProjectFile(data)
		if err != nil {
			return nil, err
		}

		projectFile.Content = contents
		files = append(files, projectFile)
	}

	for _, phase := range data.Project.Phases {
		if model.HasInputs(phase) || model.HasOutputs(phase) {
			for _, phaseFile := range phaseFiles(data, phase) {
				contents, err := phaseFile.GenPhaseFile(data, phase)
				if err != nil {
					return nil, err
				}
				phaseFile.Content = contents
				files = append(files, phaseFile)
			}
		}
	}

	return files, nil
}

func renderProjectFile(data *model.TemplateData, templateFile string) (string, error) {
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

func renderPhaseFile(data *model.TemplateData, phase model.Phase, templateFile string) (string, error) {
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
