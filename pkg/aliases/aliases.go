// This package is a convenience wrapper for commonly used k8s resources and CRDs

package aliases

import (
	trafficsplitv1alpha2 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha2"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	Pod = corev1.Pod
	ConfigMap = corev1.ConfigMap
	Deployment = appsv1.Deployment
	Service = corev1.Service
	TrafficSplit = trafficsplitv1alpha2.TrafficSplit
	ReplicaSet = appsv1.ReplicaSet
	VirtualService = istiov1alpha3.VirtualService

	ConfigMaps = []*ConfigMap
	Pods = []*Pod
	Deployments = []*Deployment
	Services = []*Service
	TrafficSplits = []*TrafficSplit
	ReplicaSets = []*ReplicaSet
	VirtualServices = []*VirtualService
)


var (
	SchemeBuilder = runtime.SchemeBuilder{}
)

func init() {
	SchemeBuilder = append(SchemeBuilder, istiov1alpha3.AddToScheme)
	SchemeBuilder = append(SchemeBuilder, trafficsplitv1alpha2.AddToScheme)
}

func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
