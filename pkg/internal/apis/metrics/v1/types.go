package v1

import (
	"encoding/json"

	"github.com/prometheus/common/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetricValue is the Schema for the MetricValue API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=MetricValue,scope=Namespaced
type MetricValue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Value *Value `json:"value,omitempty"`
}

type Value struct {
	model.Value
}

func (in *Value) DeepCopyInto(out *Value) {
	var o Value
	b, err := json.Marshal(in)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &o)
	if err != nil {
		return
	}
	*out = o
	return
}

func (in *Value) DeepCopyValue() *Value {
	if in == nil {
		return nil
	}
	out := new(Value)
	in.DeepCopyInto(out)
	return out
}

// MetricValueList contains a list of MetricValues
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MetricValueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetricValue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetricValue{}, &MetricValueList{})
}
