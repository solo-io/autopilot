package templates

import (
	"github.com/solo-io/autopilot/codegen/model"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateFunc func(data *model.ProjectData) runtime.Object
