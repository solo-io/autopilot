package render

import (
	"bytes"
	"sort"
	"text/template"

	"github.com/gobuffalo/packr"
)

// exported interface for using to render templates
type TemplateRenderer interface {
	ExecuteTemplate(name, templateText string, data interface{}) (string, error)
}

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
	// core templates, baked into the binary
	templates packr.Box

	// custom delimiters to use for evaluations
	left, right string
}

var DefaultTemplateRenderer = templateRenderer{
	// a packr.Box for reading the conents of ../templates
	// note that this code uses relative path
	// and will need to be updated if the relative
	// path from this file to the templates dir changes
	templates: packr.NewBox("../templates"),
}

func (r templateRenderer) renderCoreTemplates(templatesToRender inputTemplates, data interface{}) ([]OutFile, error) {
	var renderedFiles []OutFile
	for tmplPath, out := range templatesToRender {
		content, err := r.renderCoreTemplate(tmplPath, data)
		if err != nil {
			return nil, err
		}
		out.Content = content
		renderedFiles = append(renderedFiles, out)
	}
	return renderedFiles, nil
}

func (r templateRenderer) renderCoreTemplate(path string, data interface{}) (string, error) {
	templateText, err := r.templates.FindString(path)
	if err != nil {
		return "", err
	}

	return r.executeTemplate(path, templateText, data)
}

func (r templateRenderer) renderCustomTemplates(customTemplates map[string]string, data interface{}) ([]OutFile, error) {
	var renderedFiles []OutFile
	for outPath, templateText := range customTemplates {
		content, err := r.executeTemplate(outPath, templateText, data)
		if err != nil {
			return nil, err
		}
		out := OutFile{
			Path:    outPath,
			Content: content,
		}
		renderedFiles = append(renderedFiles, out)
	}
	sort.SliceStable(renderedFiles, func(i, j int) bool {
		return renderedFiles[i].Path < renderedFiles[j].Path
	})
	return renderedFiles, nil
}

func (r templateRenderer) ExecuteTemplate(name, templateText string, data interface{}) (string, error) {
	return r.executeTemplate(name, templateText, data)
}

func (r templateRenderer) executeTemplate(name, templateText string, data interface{}) (string, error) {

	funcs := makeTemplateFuncs()

	tmpl := template.New(name).Funcs(funcs)

	if r.left != "" {
		tmpl = tmpl.Delims(r.left, r.right)
	}

	tmpl, err := tmpl.Parse(templateText)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Funcs(funcs).Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
