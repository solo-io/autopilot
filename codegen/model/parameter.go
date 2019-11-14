package model

import (
	"reflect"

	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/autopilot/api/v1"
)

type Parameter v1.Parameter

func (p Parameter) String() string {
	return string(p.LowerName)
}

func (p Parameter) Equals(parameter Parameter) bool {
	return reflect.DeepEqual(p, parameter)
}

// registered parameters
var Parameters []Parameter

// register the proto version of the parameter
func Register(param v1.Parameter) Parameter {
	return RegisterModel(Parameter(param))
}

// register the parameter for code generation
func RegisterModel(param Parameter) Parameter {
	for _, existing := range Parameters {
		if existing.LowerName == param.LowerName {
			logrus.Fatalf("parameter %v already defined", param.LowerName)
		}
	}
	Parameters = append(Parameters, param)
	return param
}

var (
	Events = RegisterModel(Parameter{
		LowerName:    "events",
		PluralName:   "Events",
		SingleName:   "Event",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	ConfigMaps = RegisterModel(Parameter{
		LowerName:    "configmaps",
		PluralName:   "ConfigMaps",
		SingleName:   "ConfigMap",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	Services = RegisterModel(Parameter{
		LowerName:    "services",
		PluralName:   "Services",
		SingleName:   "Service",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	Pods = RegisterModel(Parameter{
		LowerName:    "pods",
		PluralName:   "Pods",
		SingleName:   "Pod",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	Jobs = RegisterModel(Parameter{
		LowerName:    "jobs",
		PluralName:   "Jobs",
		SingleName:   "Job",
		ImportPrefix: "batchv1",
		Package:      "k8s.io/api/batch/v1",
		ApiGroup:     "batch",
	})
	Deployments = RegisterModel(Parameter{
		LowerName:    "deployments",
		PluralName:   "Deployments",
		SingleName:   "Deployment",
		ImportPrefix: "appsv1",
		Package:      "k8s.io/api/apps/v1",
		ApiGroup:     "apps",
	})
	ReplicaSets = RegisterModel(Parameter{
		LowerName:    "replicasets",
		PluralName:   "ReplicaSets",
		SingleName:   "ReplicaSet",
		ImportPrefix: "appsv1",
		Package:      "k8s.io/api/apps/v1",
		ApiGroup:     "apps",
	})
	VirtualServices = RegisterModel(Parameter{
		LowerName:    "virtualservices",
		PluralName:   "VirtualServices",
		SingleName:   "VirtualService",
		ImportPrefix: "istiov1alpha3",
		Package:      "istio.io/client-go/pkg/apis/networking/v1alpha3",
		ApiGroup:     "networking.istio.io",
		IsCrd:        true,
	})
	Gateways = RegisterModel(Parameter{
		LowerName:    "gateways",
		PluralName:   "Gateways",
		SingleName:   "Gateway",
		ImportPrefix: "istiov1alpha3",
		Package:      "istio.io/client-go/pkg/apis/networking/v1alpha3",
		ApiGroup:     "networking.istio.io",
		IsCrd:        true,
	})
	Metrics = RegisterModel(Parameter{
		LowerName:    "metrics",
		PluralName:   "Metrics",
		SingleName:   "Metric",
		ImportPrefix: "metrics",
		Package:      "github.com/solo-io/autopilot/pkg/metrics",
	})
)
