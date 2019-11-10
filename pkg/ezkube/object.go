package ezkube

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// the EzKube Object is a wrapper for a kubernetes runtime.Object
// which contains Kubernetes metadata
type Object interface {
	runtime.Object
	v1.Object
}

// the EzKube Object is a wrapper for a kubernetes List object
type List interface {
	runtime.Object
	v1.ListInterface
}
