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

type GenFile struct {
	OutPath       string
	TemplatePath  string
	SkipOverwrite bool
	Content       string
}

func projectFiles(data *TemplateData) []GenFile {
	return []GenFile{
		{OutPath: filepath.Join(data.ProjectPackage, "main.go"), TemplatePath: "main.gotmpl"},
		{OutPath: filepath.Join(data.SchedulerImportPath, "scheduler.go"), TemplatePath: "scheduler.gotmpl"},
		{OutPath: filepath.Join(data.ConfigImportPath, "config.go"), TemplatePath: "config.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "doc.go"), TemplatePath: "doc.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "phases.go"), TemplatePath: "phases.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "register.go"), TemplatePath: "register.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "spec.go"), TemplatePath: "spec.gotmpl"},
		{OutPath: filepath.Join(data.TypesImportPath, "types.go"), TemplatePath: "types.gotmpl"},
	}
}
func phaseFiles(data *TemplateData, phase Phase) []GenFile {
	return []GenFile{
		{OutPath: filepath.Join(data.ProjectPackage, "pkg", "workers", workerImportPrefix(phase), "parameters.go"), TemplatePath: "parameters.gotmpl"},
		{OutPath: filepath.Join(data.ProjectPackage, "pkg", "workers", workerImportPrefix(phase), "worker.go"), TemplatePath: "worker.gotmpl", SkipOverwrite: true},
	}
}

func Generate(data *TemplateData) ([]GenFile, error) {
	var files []GenFile
	for _, projectFile := range projectFiles(data) {
		contents, err := renderProjectFile(data, projectFile.TemplatePath)
		if err != nil {
			return nil, err
		}
		projectFile.Content = contents
		files = append(files, projectFile)
	}

	for _, phase := range data.Project.Phases {
		if hasInputs(phase) || hasOutputs(phase) {
			for _, phaseFile := range phaseFiles(data, phase) {
				contents, err := renderWorkerFile(data, phase, phaseFile.TemplatePath)
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
