// Definitions for the Kubernetes types
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Paint is the Schema for the paint API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=paint,scope=Namespaced
type Paint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec        PaintColor  `json:"spec,omitempty"`
	PaintStatus PaintStatus `json:"status,omitempty"`
}

// PaintStatus defines an observed condition of Paint
// +k8s:openapi-gen=true
type PaintStatus struct {
	// ObservedGeneration was the last metadata.generation of the Paint
	// observed by the operator. If this does not match the metadata.generation of the Paint,
	// it means the operator has not yet reconciled the current generation of the operator
	ObservedGeneration int64 `json:"observedGeneration",omitempty`

	// StatusInfo defines the observed state of the Paint in the cluster
	TubeStatus
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
