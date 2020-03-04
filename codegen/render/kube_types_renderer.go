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
	"code/types/clients.gotmpl": {
		Path: "clients.go",
	},
	"code/types/register.gotmpl": {
		Path: "register.go",
	},
	"code/types/doc.gotmpl": {
		Path: "doc.go",
	},
}

var controllerTemplates = inputTemplates{
	"code/controller/event_handlers.gotmpl": {
		Path: "controller/event_handlers.go",
	},
	"code/controller/reconcilers.gotmpl": {
		Path: "controller/reconcilers.go",
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

	files, err := r.renderCoreTemplates(templatesToRender, grp)
	if err != nil {
		return nil, err
	}

	customFiles, err := r.renderCustomTemplates(grp.CustomTemplates, grp)
	if err != nil {
		return nil, err
	}

	// prepend output file paths with path to api dir
	for i, out := range append(files, customFiles...) {
		out.Path = filepath.Join(r.ApiRoot, grp.Group, grp.Version, out.Path)
		files[i] = out
	}

	return files, nil
}
