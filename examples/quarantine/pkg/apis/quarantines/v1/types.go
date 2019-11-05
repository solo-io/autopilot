package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Quarantine is the Schema for the quarantine API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=quarantine,scope=Namespaced
type Quarantine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuarantineSpec   `json:"spec,omitempty"`
	Status QuarantineStatus `json:"status,omitempty"`
}

// QuarantineStatus defines an observed condition of Quarantine
// +k8s:openapi-gen=true
type QuarantineStatus struct {
	// Phase is a required field. Do not modify it directly
	Phase QuarantinePhase `json:"phase,omitempty"`

	// StatusInfo defines the observed state of the cluster
	QuarantineStatusInfo
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// QuarantineList contains a list of Quarantine
type QuarantineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Quarantine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Quarantine{}, &QuarantineList{})
}
