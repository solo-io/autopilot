package v1

// EDIT THIS FILE!  This file should contain the definitions for your API Spec and Status!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "autopilot generate" to regenerate code after modifying this file
// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

// QuarantineSpec defines the desired state of Quarantine
// +k8s:openapi-gen=true
type QuarantineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
}

// QuarantineStatusInfo defines an observed condition of Quarantine
// +k8s:openapi-gen=true
type QuarantineStatusInfo struct {
	// INSERT ADDITIONAL STATUS FIELDS - observed state of cluster
}
