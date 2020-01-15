package render

import (
	"bytes"
	"github.com/gobuffalo/packr"
	"text/template"
)

// map of template files to the file they render to
type inputTemplates map[string]OutFile

func (r *inputTemplates) add(t inputTemplates) {
	if *r == nil {
		*r = make(map[string]OutFile)
	}
	for file, out := range t {
		(*r)[file] = out
	}
}

type templateRenderer struct {
	// templates baked into the binary
	templates packr.Box

	// custom delimiters to use for evaluations
	left, right string
}

var defaultTemplateRenderer = templateRenderer{
	// a packr.Box for reading the conents of ../templates
	// note that this code uses relative path
	// and will need to be updated if the relative
	// path from this file to the templates dir changes
	templates: packr.NewBox("../templates"),
}

func (r templateRenderer) renderInputs(templatesToRender inputTemplates, data interface{}) ([]OutFile, error) {
	var renderedFiles []OutFile
	for tmplPath, out := range templatesToRender {

		content, err := r.renderFile(tmplPath, data)
		if err != nil {
			return nil, err
		}
		out.Content = content
		renderedFiles = append(renderedFiles, out)
	}
	return renderedFiles, nil
}

func (r templateRenderer) renderFile(path string, data interface{}) (string, error) {
	templateText, err := r.templates.FindString(path)
	if err != nil {
		return "", err
	}

	funcs := makeTemplateFuncs()

	tmpl := template.New(path).Funcs(funcs)

	if r.left != "" {
		tmpl = tmpl.Delims(r.left, r.right)
	}

	tmpl, err = tmpl.Parse(templateText)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Funcs(funcs).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
