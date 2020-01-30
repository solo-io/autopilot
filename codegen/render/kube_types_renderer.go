package render

import (
	"path/filepath"
	"strings"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
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

	deepCopyLocalFiles, err := r.deepCopyGenTemplate(grp)
	if err != nil {
		return nil, err
	}
	files = append(files, deepCopyLocalFiles...)

	// prepend output file paths with path to api dir
	for i, out := range files {
		out.Path = filepath.Join(r.ApiRoot, grp.Group, grp.Version, out.Path)
		files[i] = out
	}

	return files, nil
}

const (
	protoDeepCopyTemplate = "code/types/proto_deepcopy.gotmpl"
	protoDeepCopyGo       = "proto_deepcopy.go"
)

type DescriptorsWithGopath struct {
	Descriptors      []*model.DescriptorWithPath
	Resources        []Resource
	PackageName      string
	goPackageToMatch string
}

func (r KubeCodeRenderer) deepCopyGenTemplate(grp Group) ([]OutFile, error) {
	var result []OutFile
	for _, v := range uniquePackages(grp) {
		outputDir := filepath.Join(grp.Module, grp.ApiRoot, grp.GroupVersion.String())
		if v == "" {
			inputTmpls := inputTemplates{
				protoDeepCopyTemplate: OutFile{
					Path: protoDeepCopyGo,
				},
			}
			files, err := r.renderInputs(inputTmpls, DescriptorsWithGopath{
				Descriptors:      grp.Descriptors,
				PackageName:      grp.Version,
				Resources:        grp.Resources,
				goPackageToMatch: outputDir,
			})
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
		} else {
			inputTmpls := inputTemplates{
				protoDeepCopyTemplate: OutFile{
					Path: filepath.Join(strings.TrimPrefix(v, outputDir), protoDeepCopyGo),
				},
			}
			files, err := r.renderInputs(inputTmpls, DescriptorsWithGopath{
				Descriptors:      grp.Descriptors,
				PackageName:      filepath.Base(v),
				Resources:        grp.Resources,
				goPackageToMatch: v,
			})
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
		}
	}
	return result, nil
}

func uniquePackages(grp Group) []string {
	resultMap := make(map[string]struct{})
	for _, v := range grp.Resources {
		resultMap[v.Package] = struct{}{}
	}
	var result []string
	for k, _ := range resultMap {
		result = append(result, k)
	}
	return result
}

