package render

import (
	"strings"
	"text/template"

	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/iancoleman/strcase"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"sigs.k8s.io/yaml"
)

func makeTemplateFuncs() template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add some extra functionality
	extra := template.FuncMap{
		// string utils

		"toToml":   toTOML,
		"toYaml":   toYAML,
		"fromYaml": fromYAML,
		"toJson":   toJSON,
		"fromJson": fromJSON,

		"join":            strings.Join,
		"lower":           strings.ToLower,
		"lower_camel":     strcase.ToLowerCamel,
		"upper_camel":     strcase.ToCamel,
		"snake":           strcase.ToSnake,
		"split":           splitTrimEmpty,
		"string_contains": strings.Contains,

		// autopilot funcs
		"group_import_path": func(grp Group) string {
			return util.GoPackage(grp)
		},
		// Used by types.go to get all unique external imports for a groups resources
		"imports_for_group": func(grp Group) []string {
			return uniquePackages(grp)
		},
		/*
			Used by the proto_deepcopy.gotml file to decide which objects need a proto.clone deepcopy method.

			In order to support external go packages generated from protos, this template is run for
			every unique external go package, and then filters out only the protos which are relevant to
			that specific go package before generating.
		*/
		"needs_deepcopy": func(grp DescriptorsWithGopath) []*descriptor.DescriptorProto {
			uniqueFile := getUniqueRelevantDescriptorsForGroup(grp)
			var result []*descriptor.DescriptorProto
			for _, file := range uniqueFile {
				for _, desc := range file.GetMessageType() {
					result = append(result, fineDeepCopyFields(file.GetPackage(), grp.Resources, desc)...)
				}
			}
			return result
		},
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

// find the proto messages for a given set of descriptors which need proto_deepcopoy funcs
func fineDeepCopyFields(packageName string, resources []Resource, desc *descriptor.DescriptorProto) []*descriptor.DescriptorProto {
	var result []*descriptor.DescriptorProto
	var shouldGenerate bool
	for _, v := range desc.GetField() {
		if v.TypeName != nil && !strings.Contains(v.GetTypeName(), packageName) {
			shouldGenerate = true
			break
		}
	}
	if !shouldGenerate {
		for _, resource := range resources {
			if resource.Spec.Type == desc.GetName() ||
				(resource.Status != nil && resource.Status.Type == desc.GetName()) {
				shouldGenerate = true
			}
		}
	}

	if len(desc.GetOneofDecl()) > 0 || shouldGenerate {
		result = append(result, desc)
	}
	return result
}

/*
	Get the relevant descriptors for a group of descriptors with a go package to match against.
	A unique object is initialized for each external go package to the group package
*/
func getUniqueRelevantDescriptorsForGroup(grp DescriptorsWithGopath) []*model.DescriptorWithPath {
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

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// fromYAML converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromYAML(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// toTOML takes an interface, marshals it to toml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toTOML(v interface{}) string {
	b := bytes.NewBuffer(nil)
	e := toml.NewEncoder(b)
	err := e.Encode(v)
	if err != nil {
		return err.Error()
	}
	return b.String()
}

// toJSON takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

// fromJSON converts a JSON document into a map[string]interface{}.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromJSON(str string) map[string]interface{} {
	m := make(map[string]interface{})

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

func splitTrimEmpty(s, sep string) []string {
	return strings.Split(strings.TrimSpace(s), sep)
}
