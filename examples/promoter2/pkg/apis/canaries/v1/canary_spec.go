package v1

import (
	hpav1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO(User): Edit this file!!

// CanaryConditionType is the type of a CanaryCondition
type CanaryConditionType string

const (
	// PromotedType refers to the result of the last canary analysis
	PromotedType CanaryConditionType = "Promoted"
)

// CanaryCondition is a status condition for a Canary
type CanaryCondition struct {
	// Type of this condition
	Type CanaryConditionType `json:"type"`

	// Status of this condition
	Status corev1.ConditionStatus `json:"status"`

	// LastUpdateTime of this condition
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// LastTransitionTime of this condition
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason for the current status of this condition
	Reason string `json:"reason,omitempty"`

	// Message associated with this condition
	Message string `json:"message,omitempty"`
}

// CanaryPhase is a label for the condition of a canary at the current time
type CanaryPhase string

const (
	// CanaryPhaseInitializing means the canary initializing is underway
	CanaryPhaseInitializing CanaryPhase = "Initializing"
	// CanaryPhaseProgressing means the canary analysis is underway
	CanaryPhaseProgressing CanaryPhase = "Progressing"
	// CanaryPhasePromoting means the canary analysis is finished and the primary spec has been updated
	CanaryPhasePromoting CanaryPhase = "Promoting"
	// CanaryPhaseProgressing means the canary promotion is finished and traffic has been routed back to primary
	CanaryPhaseFinalising CanaryPhase = "Finalising"
	// CanaryPhaseSucceeded means the canary analysis has been successful
	// and the canary deployment has been promoted
	CanaryPhaseSucceeded CanaryPhase = "Succeeded"
	// CanaryPhaseFailed means the canary analysis failed
	// and the canary deployment has been scaled to zero
	CanaryPhaseFailed CanaryPhase = "Failed"
)

// CanaryStatus is used for state persistence (read-only)
type CanaryStatus struct {
	Phase        CanaryPhase `json:"phase"`
	FailedChecks int         `json:"failedChecks"`
	CanaryWeight int         `json:"canaryWeight"`
	Iterations   int         `json:"iterations"`

	// +optional
	LastAppliedSpec string `json:"lastAppliedSpec,omitempty"`
	// +optional
	LastPromotedSpec string `json:"lastPromotedSpec,omitempty"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Conditions []CanaryCondition `json:"conditions,omitempty"`
}

type CanarySpec struct {
	// reference to target resource
	TargetRef hpav1.CrossVersionObjectReference `json:"targetRef"`

	// virtual service spec
	Service CanaryService `json:"service"`

	// metrics and thresholds
	CanaryAnalysis CanaryAnalysis `json:"canaryAnalysis"`
}

// CanaryService is used to create ClusterIP services
// and Istio Virtual Service
type CanaryService struct {
	Port          int32              `json:"port"`
	PortName      string             `json:"portName,omitempty"`
	TargetPort    intstr.IntOrString `json:"targetPort,omitempty"`
	PortDiscovery bool               `json:"portDiscovery"`
	Timeout       string             `json:"timeout,omitempty"`
	//// Istio
	//Gateways      []string                         `json:"gateways,omitempty"`
	//Hosts         []string                         `json:"hosts,omitempty"`
	//TrafficPolicy *istiov1alpha3.TrafficPolicy     `json:"trafficPolicy,omitempty"`
	//Match         []istiov1alpha3.HTTPMatchRequest `json:"match,omitempty"`
	//Rewrite       *istiov1alpha3.HTTPRewrite       `json:"rewrite,omitempty"`
	//Retries       *istiov1alpha3.HTTPRetry         `json:"retries,omitempty"`
	//Headers       *istiov1alpha3.Headers           `json:"headers,omitempty"`
	//CorsPolicy    *istiov1alpha3.CorsPolicy        `json:"corsPolicy,omitempty"`
	//// App Mesh
	//MeshName string   `json:"meshName,omitempty"`
	//Backends []string `json:"backends,omitempty"`
}

// CanaryAnalysis is used to describe how the analysis should be done
type CanaryAnalysis struct {
	Interval   string         `json:"interval"`
	Threshold  int            `json:"threshold"`
	MaxWeight  int            `json:"maxWeight"`
	StepWeight int            `json:"stepWeight"`
	Metrics    []CanaryMetric `json:"metrics"`
	//Webhooks   []CanaryWebhook                  `json:"webhooks,omitempty"`
	//Match      []istiov1alpha3.HTTPMatchRequest `json:"match,omitempty"`
	Iterations int `json:"iterations,omitempty"`
}

// CanaryMetric holds the reference to Istio metrics used for canary analysis
type CanaryMetric struct {
	Name      string  `json:"name"`
	Interval  string  `json:"interval,omitempty"`
	Threshold float64 `json:"threshold"`
	// +optional
	Query string `json:"query,omitempty"`
}
