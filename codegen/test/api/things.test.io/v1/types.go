// Definitions for the Kubernetes types
package v1

import (
	// . "github.com/solo-io/autopilot/codegen/test/api/things.test.io/v1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status

// Paint is the Schema for the paint API
type Paint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PaintSpec   `json:"spec,omitempty"`
	Status PaintStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PaintList contains a list of Paint
type PaintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Paint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Paint{}, &PaintList{})
}
