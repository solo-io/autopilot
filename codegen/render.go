package codegen

import (
	"bytes"
	"github.com/gobuffalo/packr"
	"github.com/solo-io/autopilot/codegen/templates"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"text/template"
)

type OutFile struct {
	Path          string
	SkipOverwrite bool
	Permission    os.FileMode
	Content       string // set by Renderer
}

// renders generated code from templates
type Renderer struct {
	Templates     packr.Box
	Funcs         template.FuncMap
	ResourcePaths map[string]OutFile
}

type Group struct {
	schema.GroupVersion
	Resources []Resource
}

type Resource struct {
	Group  // the group I belong to
	Kind   string
	Spec   Field
	Status *Field
}

type Field struct {
	Type string
}

var defaultRenderer = Renderer{
	Templates: packr.NewBox("./templates"),
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
	},
}

func RenderGroup(grp Group) ([]OutFile, error) {
	return defaultRenderer.RenderGroup(grp)
}

func (r Renderer) RenderGroup(grp Group) ([]OutFile, error) {
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

func (r Renderer) renderFile(path string, data interface{}) (string, error) {
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
