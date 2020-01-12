package render

import (
	"bytes"
	"text/template"

	"github.com/gobuffalo/packr"
)

// map of template files to the file they render to
type resourceTemplates map[string]OutFile

func (r *resourceTemplates) add(file string, out OutFile) {
	if *r == nil {
		*r = make(map[string]OutFile)
	}
	(*r)[file] = out
}

// a packr.Box for reading the conents of ../templates
// note that this code uses relative path
// and will need to be updated if the relative
// path from this file to the templates dir changes
var templateBox = packr.NewBox("../templates")

// renders kubernetes from templates
type KubeCodeRenderer struct {
	// templates baked into the binary
	templates packr.Box

	// the templates to use for rendering kube kypes
	TypesTemplates resourceTemplates
	// the templates to use for rendering kube controllers
	ControllerTemplates resourceTemplates

	// the go module of the project
	GoModule string

	// the relative path to the api dir
	// types will render in the package <module>/<apiRoot>/<group>/<version>
	ApiRoot string
}

var typesTemplates = resourceTemplates{
	"code/types/types.gotmpl": {
		Path: "types.go",
	},
	"code/types/register.gotmpl": {
		Path: "register.go",
	},
	"code/types/doc.gotmpl": {
		Path: "doc.go",
	},
	"code/types/proto_deepcopy.gotmpl": {
		Path: "proto_deepcopy.go",
	},
}

var controllerTemplates = resourceTemplates{
	"code/controller/controller.gotmpl": {
		Path: "controller/controller.go",
	},
}

func RenderApiTypes(grp Group) ([]OutFile, error) {
	defaultKubeCodeRenderer := KubeCodeRenderer{
		templates:           templateBox,
		TypesTemplates:      typesTemplates,
		ControllerTemplates: controllerTemplates,
		GoModule:            grp.Module,
		ApiRoot:             grp.ApiRoot,
	}

	return defaultKubeCodeRenderer.RenderKubeCode(grp)
}

func (r KubeCodeRenderer) RenderKubeCode(grp Group) ([]OutFile, error) {
	templatesToRender := make(resourceTemplates)
	if grp.RenderTypes {
		for tmplPath, out := range r.TypesTemplates {
			templatesToRender.add(tmplPath, out)
		}
	}
	if grp.RenderController {
		for tmplPath, out := range r.ControllerTemplates {
			templatesToRender.add(tmplPath, out)
		}
	}

	var renderedFiles []OutFile
	for tmplPath, out := range templatesToRender {

		content, err := r.renderFile(tmplPath, grp)
		if err != nil {
			return nil, err
		}
		out.Content = content
		// prepend with version dir
		out.Path = r.ApiRoot + "/" + grp.Group + "/" + grp.Version + "/" + out.Path
		renderedFiles = append(renderedFiles, out)
	}
	return renderedFiles, nil
}

func (r KubeCodeRenderer) renderFile(path string, data interface{}) (string, error) {
	templateText, err := r.templates.FindString(path)
	if err != nil {
		return "", err
	}

	funcs := makeTemplateFuncs(r.GoModule, r.ApiRoot)

	tmpl, err := template.New(path).Funcs(funcs).Parse(templateText)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Funcs(funcs).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
