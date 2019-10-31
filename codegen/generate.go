package codegen

import (
	"bytes"
	"fmt"
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

func Load(file string) (*TemplateData, error) {
	projData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var project Project
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

	data := &TemplateData{
		Project:             project,
		ProjectPackage:      projectGoPkg,
		Group:               apiGroup,
		Version:             apiVersion,
		TypesImportPath:     apiImportPath,
		SchedulerImportPath: schedulerImportPath,
		ConfigImportPath:    configImportPath,
		KindLowerCamel:      strcase.ToLowerCamel(project.Kind),
	}

	// required for use by worker template
	for i, phase := range project.Phases {
		phase.Project = data
		project.Phases[i] = phase
	}

	return data, nil
}

func projectFiles(data *TemplateData) map[string]string {
	return map[string]string{
		filepath.Join(data.ProjectPackage, "main.go"):           "main.gotmpl",
		filepath.Join(data.SchedulerImportPath, "scheduler.go"): "scheduler.gotmpl",
		filepath.Join(data.ConfigImportPath, "config.go"):       "config.gotmpl",
		filepath.Join(data.TypesImportPath, "doc.go"):           "doc.gotmpl",
		filepath.Join(data.TypesImportPath, "phases.go"):        "phases.gotmpl",
		filepath.Join(data.TypesImportPath, "register.go"):      "register.gotmpl",
		filepath.Join(data.TypesImportPath, "spec.go"):          "spec.gotmpl",
		filepath.Join(data.TypesImportPath, "types.go"):         "types.gotmpl",
	}
}
func phaseFiles(data *TemplateData, phase Phase) map[string]string {
	return map[string]string{
		filepath.Join(data.ProjectPackage, "pkg", "workers", workerImportPrefix(phase), "parameters.go"): "parameters.gotmpl",
		filepath.Join(data.ProjectPackage, "pkg", "workers", workerImportPrefix(phase), "worker.go"):     "worker.gotmpl",
	}
}

func Generate(data *TemplateData) (map[string]string, error) {
	files := make(map[string]string)
	for path, projectFile := range projectFiles(data) {
		contents, err := renderProjectFile(data, projectFile)
		if err != nil {
			return nil, err
		}
		files[path] = contents
	}

	for _, phase := range data.Project.Phases {
		if hasInputs(phase) || hasOutputs(phase) {
			for path, phaseFile := range phaseFiles(data, phase) {
				contents, err := renderWorkerFile(data, phase, phaseFile)
				if err != nil {
					return nil, err
				}
				files[path] = contents
			}
		}
	}

	return files, nil
}

func renderProjectFile(data *TemplateData, templateFile string) (string, error) {
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

func renderWorkerFile(data *TemplateData, phase Phase, templateFile string) (string, error) {
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
