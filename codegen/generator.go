package codegen

import (
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/autopilot/codegen/templates"
	"os"
)

// IfExistsAction determines what to do if the scaffold file already exists
type IfExistsAction int

const (
	// Skip skips the file and moves to the next one
	Skip IfExistsAction = iota

	// Error returns an error and stops processing
	Error

	// Overwrite truncates and overwrites the existing file
	Overwrite
)

// Input is the input for scaffolding a file
type File struct {
	// Path is the file to write, relative to autopilot.yaml (project root)
	Path string

	// Content is the content to write
	Content string

	// The mode/permissions to set on the file
	FileMode os.FileMode

	// IfExistsAction determines what to do if the file exists
	IfExistsAction IfExistsAction
}

// the generator generates the files for the operator
type OperatorScaffolding struct {
}

func (s OperatorScaffolding) Files(data *model.ProjectData) ([]File, error) {
	// top level files

}
