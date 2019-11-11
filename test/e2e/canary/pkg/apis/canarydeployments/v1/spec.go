package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  This file should contain the definitions for your API Spec and Status!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "autopilot generate" to regenerate code after modifying this file
// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

// CanaryDeploymentSpec defines the desired state of CanaryDeployment
// +k8s:openapi-gen=true
type CanaryDeploymentSpec struct {
	// the deployment for which to deploy canaries
	// every deployment created via CanaryDeployment will have 2 deployments made: a primary and a canary
	// when the CanaryDeployment is created, only the primary deployment will be created.
	// whenever this spec changes, the changed spec will be applied to the canary and the canary will be scaled up
	// if the promotion succeeds, the spec is applied to the primary deployment
	// if the promotion fails, the spec is not applied to the primary deployment
	// the canary is then scaled down to 0
	Deployment string `json:"deployment"`

	// ports for which traffic should be split (between primary and canary)
	Ports []int32 `json:"ports,omitempty"`

	// Canary must maintain this success rate or for the given analysisPeriod
	SuccessThreshold float64 `json:"successThreshold,omitempty"`

	// How long should we process the canary for before promoting?
	AnalysisPeriod metav1.Duration `json:"analysisPeriod,omitempty"`
}

// CanaryDeploymentStatusInfo defines an observed condition of CanaryDeployment
// +k8s:openapi-gen=true
type CanaryDeploymentStatusInfo struct {
	// used to record the time processing started
	TimeStarted metav1.Time `json:"timeStarted,omitempty"`

	// record the history of the canary (promotions/rollbacks)
	History []CanaryResult `json:"history,omitempty"`
}

// CanaryResult is either a Rollback or Promotion, along with the observed generation of the canary
// +k8s:openapi-gen=true
type CanaryResult struct {
	// if true, the canary was promoted. if false, it was rolled back
	PromotionSucceeded bool `json:"promotionSucceeded"`

	// the observed generation of the canary that was promoted or rolled back
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
}
