package render

import (
	"path/filepath"
)

// renders kubernetes from templates
type KubeCodeRenderer struct {
	templateRenderer

	// the templates to use for rendering kube kypes
	TypesTemplates inputTemplates

	// the templates to use for rendering kube controllers
	ControllerTemplates inputTemplates

	// the go module of the project
	GoModule string

	// the relative path to the api dir
	// types will render in the package <module>/<apiRoot>/<group>/<version>
	ApiRoot string
}

var typesTemplates = inputTemplates{
	"code/types/types.gotmpl": {
		Path: "types.go",
	},
	"code/types/register.gotmpl": {
		Path: "register.go",
	},
	"code/types/doc.gotmpl": {
		Path: "doc.go",
	},
}

var controllerTemplates = inputTemplates{
	"code/controller/controller.gotmpl": {
		Path: "controller/controller.go",
	},
}

func RenderApiTypes(grp Group) ([]OutFile, error) {
	defaultKubeCodeRenderer := KubeCodeRenderer{
		templateRenderer:    defaultTemplateRenderer,
		TypesTemplates:      typesTemplates,
		ControllerTemplates: controllerTemplates,
		GoModule:            grp.Module,
		ApiRoot:             grp.ApiRoot,
	}

	return defaultKubeCodeRenderer.RenderKubeCode(grp)
}

func (r KubeCodeRenderer) RenderKubeCode(grp Group) ([]OutFile, error) {
	templatesToRender := make(inputTemplates)
	if grp.RenderTypes {
		templatesToRender.add(r.TypesTemplates)
	}
	if grp.RenderController {
		templatesToRender.add(r.ControllerTemplates)
	}

	files, err := r.renderInputs(templatesToRender, grp)
	if err != nil {
		return nil, err
	}

	// prepend output file paths with path to api dir
	for i, out := range files {
		out.Path = filepath.Join(r.ApiRoot, grp.Group, grp.Version, out.Path)
		files[i] = out
	}

	return files, nil
}
