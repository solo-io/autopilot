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

	apiVersion := apiVersionParts[1]
	apiImportPath := filepath.Join(projectGoPkg, "pkg", "apis", strings.ToLower(c.Plural(project.Kind)), apiVersion)

	return &TemplateData{
		Project:           project,
		ProjectPackage:    projectGoPkg,
		TypesImportPrefix: apiVersion,
		TypesImportPath:   apiImportPath,
		KindLowerCamel:    strcase.ToLowerCamel(project.Kind),
	}, nil
}

func Generate(data *TemplateData) (map[string]string, error) {
	scheduler, err := renderProjectFile(data, "scheduler.gotmpl")
	if err != nil {
		return nil, err
	}

	files := map[string]string{
		filepath.Join(data.ProjectPackage, "pkg", "scheduler", "scheduler.go"): scheduler,
	}

	for _, phase := range data.Project.Phases {
		if hasInputs(phase) || hasOutputs(phase) {
			params, err := renderWorkerFile(data, phase, "parameters.gotmpl")
			if err != nil {
				return nil, err
			}
			files[filepath.Join(data.ProjectPackage, "pkg", "workers", workerImportPrefix(phase), "parameters.go")] = params
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
