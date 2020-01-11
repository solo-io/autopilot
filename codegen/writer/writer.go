package writer

import (
	"github.com/solo-io/autopilot/codegen/render"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileWriter interface {
	WriteFiles(files []render.OutFile) error
}

// writes to the filesystem
type DefaultFileWriter struct {
	Root string
}

func (w *DefaultFileWriter) WriteFiles(files []render.OutFile) error {
	for _, file := range files {
		name := filepath.Join(w.Root, file.Path)
		content := file.Content

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
	return nil
}
