package render

import (
	"path/filepath"
	"strings"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

// renders kubernetes from templates
type ProtoCodeRenderer struct {
	templateRenderer

	// the go module of the project
	GoModule string

	// the relative path to the api dir
	// types will render in the package <module>/<apiRoot>/<group>/<version>
	ApiRoot string
}

func RenderProtoTypes(grp Group) ([]OutFile, error) {
	defaultKubeCodeRenderer := ProtoCodeRenderer{
		templateRenderer: defaultTemplateRenderer,
		GoModule:         grp.Module,
		ApiRoot:          grp.ApiRoot,
	}

	return defaultKubeCodeRenderer.RenderProtoHelpers(grp)
}

func (r ProtoCodeRenderer) RenderProtoHelpers(grp Group) ([]OutFile, error) {

	if !grp.RenderProtos {
		return nil, nil
	}

	files, err := r.deepCopyGenTemplate(grp)
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

const (
	protoDeepCopyTemplate = "code/types/proto_deepcopy.gotmpl"
	protoDeepCopyGo       = "proto_deepcopy.go"
)

// helper type for rendering proto_deepcopy.go files
type DescriptorsWithGopath struct {
	// list of descriptors pulled from the group
	Descriptors []*model.DescriptorWithPath
	// list of resources pulled from the group
	Resources []Resource
	// package name used to render the package name in the go template
	PackageName string
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
func (r ProtoCodeRenderer) deepCopyGenTemplate(grp Group) ([]OutFile, error) {
	var result []OutFile
	for _, uniquePackage := range uniquePaths(grp) {
		var (
			inputTmpls       inputTemplates
			packageName      string
			goPackageToMatch string
		)
		if uniquePackage == "" {
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
					Path: filepath.Join(
						strings.TrimPrefix(uniquePackage, filepath.Join(r.ApiRoot, grp.Group, grp.Version)), protoDeepCopyGo,
					),
				},
			}
			goPackageToMatch = filepath.Join(grp.Module, uniquePackage)
			packageName = filepath.Base(goPackageToMatch)
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
*/
func uniquePackages(grp Group) []string {
	unique := uniquePaths(grp)
	var result []string
	for _, v := range unique {
		if v != "" {
			result = append(result, filepath.Join(grp.Module, v))
		}
	}
	return result
}

/*
	Get all of the unique paths for a group by checking the packages of the resources
	This list can include an empty string which corresponds to the local group
*/
func uniquePaths(grp Group) []string {
	resultMap := make(map[string]struct{})
	for _, v := range grp.Resources {
		if !grp.RenderProtos {
			continue
		}
		resultMap[v.RelativePathFromRoot] = struct{}{}
	}
	var result []string
	for k, _ := range resultMap {
		result = append(result, k)
	}
	return result
}
