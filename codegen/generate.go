package codegen

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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
	projectGoPkg := util.GetGoPkg(filepath.Dir(file))

	apiVersionParts := strings.Split(project.ApiVersion, "/")

	if len(apiVersionParts) != 2 {
		return nil, fmt.Errorf("%v must be format groupname/version", apiVersionParts)
	}

	c := pluralize.NewClient()

	apiVersion := apiVersionParts[1]
	apiImportPath := filepath.Join(projectGoPkg, "pkg", "apis", strings.ToLower(c.Plural(project.Kind)))

	return &TemplateData{
		Project:           project,
		ProjectPackage:    projectGoPkg,
		TypesImportPrefix: apiVersion,
		TypesImportPath:   apiImportPath,
		KindLowerCamel:    strcase.ToLowerCamel(project.Kind),
	}, nil
}

func Generate(data *TemplateData) (map[string]string, error) {
	scheduler, err := render(data, "scheduler.gotmpl")
	if err != nil {
		return nil, err
	}

	return map[string]string{
		filepath.Join(data.ProjectPackage, "pkg", "scheduler", "scheduler.go"): scheduler,
	}, nil
}

func render(data *TemplateData, templateFile string) (string, error) {
	tmpl := mustLoad(templateFile)
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func mustLoad(file string) *template.Template {
	fullPath := filepath.Join(autopilotRoot(), "codegen", "templates", file)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("failed to read template file %v", fullPath)
	}
	return template.Must(template.New(file).Parse(string(content)))
}

func autopilotRoot() string {
	return filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "solo-io", "autopilot")
}
