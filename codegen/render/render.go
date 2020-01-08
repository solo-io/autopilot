package render

import (
	"github.com/solo-io/autopilot/codegen/model"
	"os"
)

type Group = model.Group

type Resource = model.Resource

type Field = model.Field

type OutFile struct {
	Path       string
	Permission os.FileMode
	Content    string // set by KubeTypesRenderer
}
