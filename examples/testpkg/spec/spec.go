package spec

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/api/core/v1"
)

// +k8s:deepcopy-gen=true
type TestSpec struct {
	Target    v1.ObjectReference      `json:"target"`
	Vs        *v1alpha3.VirtualService `json:"route"`
	Threshold float64                 `json:"threshold"`
}
