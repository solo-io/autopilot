package util

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"sigs.k8s.io/yaml"
)

var JsonPbMarshaler = &jsonpb.Marshaler{}

func UnmarshalYaml(data []byte, into proto.Message) error {
	jsn, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		return err
	}

	return jsonpb.Unmarshal(bytes.NewBuffer(jsn), into)
}

func MarshalYaml(m proto.Message) ([]byte, error) {
	s, err := JsonPbMarshaler.MarshalToString(m)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(s))
}
