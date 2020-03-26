package render

import (
	"path/filepath"
)

// renders kubernetes from templates
type KubeCodeRenderer struct {
	templateRenderer

	// the templates to use for rendering kube kypes
	TypesTemplates inputTemplates

	// the templates to use for rendering typed kube clients which use the underlying cache
	ClientsTemplates inputTemplates

	// the templates to use for rendering kube controllers
	ControllerTemplates inputTemplates

	// the go module of the project
	GoModule string

	// the relative path to the api dir
	// types will render in the package <module>/<apiRoot>/<group>/<version>
	ApiRoot string
}

var TypesTemplates = inputTemplates{
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

var ClientsTemplates = inputTemplates{
	"code/types/clients.gotmpl": {
		Path: "clients.go",
	},
}

var ControllerTemplates = inputTemplates{
	"code/controller/event_handlers.gotmpl": {
		Path: "controller/event_handlers.go",
	},
	"code/controller/reconcilers.gotmpl": {
		Path: "controller/reconcilers.go",
	},
}

func RenderApiTypes(grp Group) ([]OutFile, error) {
	defaultKubeCodeRenderer := KubeCodeRenderer{
		templateRenderer:    DefaultTemplateRenderer,
		TypesTemplates:      TypesTemplates,
		ClientsTemplates:    ClientsTemplates,
		ControllerTemplates: ControllerTemplates,
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
	if grp.RenderClients {
		templatesToRender.add(r.ClientsTemplates)
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

	files = append(files, customFiles...)

	// prepend output file paths with path to api dir
	for i, out := range files {
		out.Path = filepath.Join(r.ApiRoot, grp.Group, grp.Version, out.Path)
		files[i] = out
	}

	return files, nil
}
