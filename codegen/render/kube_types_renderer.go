package render

import (
	"path/filepath"

	model2 "github.com/solo-io/autopilot/codegen/model"
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

// helper type for rendering proto_deepcopy.go files
type DescriptorsWithGopath struct {
	// list of descriptors pulled from the group
	Descriptors      []*model.DescriptorWithPath
	// list of resources pulled from the group
	Resources        []Resource
	// package name used to render the package name in the go template
	PackageName      string
	// full go package which the template render funcs will use to match against the
	// set of descriptors to find the relevant messages
	goPackageToMatch string
}

/*
	Create and render the templates for the proto_deepcopy filesin order to support
	proto_deepcopy funcs for packages which are different than the main group package

	The empty string package name is treated as local, and so it it computed the same way as before

	Any other package name is than rendered to the relative path supplied.
*/
func (r KubeCodeRenderer) deepCopyGenTemplate(grp Group) ([]OutFile, error) {
	var result []OutFile
	for _, v := range uniquePackages(grp) {
		var (
			inputTmpls       inputTemplates
			packageName      string
			goPackageToMatch string
		)
		if v.GoPackage == "" {
			inputTmpls = inputTemplates{
				protoDeepCopyTemplate: OutFile{
					Path: protoDeepCopyGo,
				},
			}
			goPackageToMatch = filepath.Join(grp.Module, grp.ApiRoot, grp.GroupVersion.String())
			packageName = grp.Version
		} else {
			inputTmpls = inputTemplates{
				protoDeepCopyTemplate: OutFile{
					Path: filepath.Join(v.RelativePath, protoDeepCopyGo),
				},
			}
			goPackageToMatch = v.GoPackage
			packageName = filepath.Base(v.GoPackage)
		}
		files, err := r.renderInputs(inputTmpls, DescriptorsWithGopath{
			Descriptors:      grp.Descriptors,
			Resources:        grp.Resources,
			PackageName:      packageName,
			goPackageToMatch: goPackageToMatch,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, files...)
	}
	return result, nil
}

/*
	Get all of the unique go packages for a group by checking the packages of the resources
	This list can include abn empty string which corresponds to the local group package
*/
func uniquePackages(grp Group) []*model2.ExternalPackage {
	resultMap := make(map[string]*model2.ExternalPackage)
	for _, v := range grp.Resources {
		if !grp.RenderProtos {
			continue
		}
		if v.Package != nil {
			resultMap[v.Package.GoPackage] = v.Package
		}
	}
	var result []*model2.ExternalPackage
	for _, v := range resultMap {
		result = append(result, v)
	}
	return result
}
