package render

import (
	"bytes"
	"github.com/gobuffalo/packr"
	"github.com/solo-io/autopilot/codegen/templates"
	"text/template"
)

// renders kubernetes from templates
type KubeTypesRenderer struct {
	Templates     packr.Box
	Funcs         template.FuncMap
	ResourcePaths map[string]OutFile
}

var defaultKubeTypesRenderer = KubeTypesRenderer{
	Templates: packr.NewBox("../templates"),
	Funcs:     templates.Funcs,
	ResourcePaths: map[string]OutFile{
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
	},
}

func RenderApiTypes(grp Group) ([]OutFile, error) {
	return defaultKubeTypesRenderer.RenderKubeTypes(grp)
}

func (r KubeTypesRenderer) RenderKubeTypes(grp Group) ([]OutFile, error) {
	var renderedFiles []OutFile
	for tmplPath, out := range r.ResourcePaths {
		content, err := r.renderFile(tmplPath, grp)
		if err != nil {
			return nil, err
		}
		out.Content = content
		// prepend with version dir
		out.Path = grp.Group + "/" + grp.Version + "/" + out.Path
		renderedFiles = append(renderedFiles, out)
	}
	return renderedFiles, nil
}

func (r KubeTypesRenderer) renderFile(path string, data interface{}) (string, error) {
	templateText, err := r.Templates.FindString(path)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(path).Funcs(r.Funcs).Parse(templateText)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Funcs(r.Funcs).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
