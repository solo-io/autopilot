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
type descriptorsWithGopath struct {
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
	Get the relevant descriptors for a group of descriptors with a go package to match against.
	A unique object is initialized for each external go package to the group package
*/
func (grp descriptorsWithGopath) getUniqueDescriptors() []*model.DescriptorWithPath {
	result := make(map[string]*model.DescriptorWithPath)
	for _, desc := range grp.Descriptors {
		if desc.GetOptions().GetGoPackage() == grp.goPackageToMatch {
			result[desc.ProtoFilePath] = desc
		}
	}
	var array []*model.DescriptorWithPath
	for _, v := range result {
		array = append(array, v)
	}
	return array
}

/*
	Create and render the templates for the proto_deepcopy filesin order to support
	proto_deepcopy funcs for packages which are different than the main group package

	The empty string package name is treated as local, and so it it computed the same way as before

	Any other package name is than rendered to the relative path supplied.
*/
func (r ProtoCodeRenderer) deepCopyGenTemplate(grp Group) ([]OutFile, error) {
	var result []OutFile
	for _, uniquePackage := range uniqueGoImportPathsForGroup(grp) {
		var (
			inputTmpls       inputTemplates
			packageName      string
			goPackageToMatch string
		)
		if uniquePackage == "" {
			// this case represents the local package, therefore no special pathing is necessary
			inputTmpls = inputTemplates{
				protoDeepCopyTemplate: OutFile{
					Path: protoDeepCopyGo,
				},
			}
			goPackageToMatch = filepath.Join(grp.Module, grp.ApiRoot, grp.GroupVersion.String())
			packageName = grp.Version
		} else {
			// this case represents any go packages which are in the same go module as the root,
			// but in a different directory
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
		files, err := r.renderInputs(inputTmpls, descriptorsWithGopath{
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
func uniqueGoPackagesForGroup(grp Group) []string {
	unique := uniqueGoImportPathsForGroup(grp)
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
	This list can include an empty string ("") which corresponds to the local group
*/
func uniqueGoImportPathsForGroup(grp Group) []string {
	resultMap := make(map[string]struct{})
	for _, v := range grp.Resources {
		// if the group does not have protos to render, than finding the go import path is unnecessary
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
