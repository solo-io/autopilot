package codegen

import (
	"github.com/solo-io/autopilot/codegen/util"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func Run(dir string) error {
	config := filepath.Join(dir, "autopilot.yaml")

	project, err := Load(config)
	if err != nil {
		return err
	}

	files, err := Generate(project)
	if err != nil {
		return err
	}

	if err := util.DeepcopyGen(project.TypesImportPath); err != nil {
		return err
	}

	for name, content := range files {
		name = filepath.Join(os.Getenv("GOPATH"), "src", name)
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(name, []byte(content), 0644); err != nil {
			return err
		}

		formatted, err := imports.Process(name, []byte(content), nil)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(name, []byte(formatted), 0644); err != nil {
			return err
		}
	}

	log.Printf("Finished generating %v", project.ApiVersion+"."+project.Kind)

	return nil
}
