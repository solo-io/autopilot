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
			return uniqueGoPackagesForGroup(grp)
		},
		/*
			Used by the proto_deepcopy.gotml file to decide which objects need a proto.clone deepcopy method.

			In order to support external go packages generated from protos, this template is run for
			every unique external go package, and then filters out only the protos which are relevant to
			that specific go package before generating.
		*/
		"needs_deepcopy": func(grp descriptorsWithGopath) []*descriptor.DescriptorProto {
			uniqueFile := grp.getUniqueDescriptors()
			var result []*descriptor.DescriptorProto
			for _, file := range uniqueFile {
				for _, desc := range file.GetMessageType() {
					// for each message, in each file, find the fields which need deep copy functions
					result = append(result, findDeepCopyFields(file.GetPackage(), grp.Resources, desc)...)
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

/*
	Find the proto messages for a given set of descriptors which need proto_deepcopoy funcs.
	The three cases are as follows:

	1. One of the subfields has an external type
	2. The descriptor is either the status or the spec of one of the group resources
	3. There is a oneof present
*/

func findDeepCopyFields(packageName string, resources []Resource, desc *descriptor.DescriptorProto) []*descriptor.DescriptorProto {
	var result []*descriptor.DescriptorProto
	var shouldGenerate bool
	// case 1 above
	for _, v := range desc.GetField() {
		if v.TypeName != nil && !strings.Contains(v.GetTypeName(), packageName) {
			shouldGenerate = true
			break
		}
	}
	if !shouldGenerate {
		// case 2 above
		for _, resource := range resources {
			if resource.Spec.Type == desc.GetName() ||
				(resource.Status != nil && resource.Status.Type == desc.GetName()) {
				shouldGenerate = true
			}
		}
	}

	// case 3 above
	if len(desc.GetOneofDecl()) > 0 || shouldGenerate {
		result = append(result, desc)
	}
	return result
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
