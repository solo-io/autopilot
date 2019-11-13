package codegen

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/autopilot/pkg/defaults"
	"golang.org/x/tools/imports"

	"os"
	"path/filepath"
	"strings"
)

func Run(dir string, forceOverwrite bool) error {
	project := filepath.Join(dir, defaults.AutoPilotFile)
	operator := filepath.Join(dir, defaults.OperatorFile)

	data, err := Load(project, operator)
	if err != nil {
		return err
	}

	if err := data.Validate(); err != nil {
		return err
	}

	files, err := Generate(data)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := filepath.Join(dir, file.OutPath)
		content := file.Content

		if !forceOverwrite && file.SkipOverwrite {
			if _, err := os.Stat(name); err == nil {
				log.Printf("skippinng file %v because it exists", name)
				continue
			}
		}

		if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
			return err
		}

		perm := file.Permission
		if perm == 0 {
			perm = 0644
		}

		log.Printf("Writing %v", name)

		if err := ioutil.WriteFile(name, []byte(content), perm); err != nil {
			return err
		}

		if !strings.HasSuffix(name, ".go") {
			continue
		}

		formatted, err := imports.Process(name, []byte(content), nil)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(name, []byte(formatted), 0644); err != nil {
			return err
		}
	}

	log.Printf("Generating Deepcopy types for %v", data.TypesImportPath)
	if err := util.DeepcopyGen(data.TypesImportPath); err != nil {
		return err
	}

	log.Printf("Finished generating %v", data.ApiVersion+"."+data.Kind)

	return nil
}
