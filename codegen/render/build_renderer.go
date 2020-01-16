package render

import "github.com/solo-io/autopilot/codegen/model"

// renders files used to build the operator
type BuildRenderer struct {
	templateRenderer
}

var defaultBuildInputs = inputTemplates{
	"build/Dockerfile.tmpl": {
		Path: "Dockerfile",
	},
}

func RenderBuild(operator model.Operator) ([]OutFile, error) {
	defaultBuildRenderer := BuildRenderer{
		templateRenderer: defaultTemplateRenderer,
	}
	return defaultBuildRenderer.Render(operator)
}

func (r BuildRenderer) Render(operator model.Operator) ([]OutFile, error) {
	templatesToRender := defaultBuildInputs

	files, err := r.renderInputs(templatesToRender, operator)
	if err != nil {
		return nil, err
	}

	return files, nil
}
