package model

import "github.com/sirupsen/logrus"

type Parameter struct {
	LowerName        string `json:"lowerName"`
	SingleName       string
	PluralName       string
	ImportPrefix     string
	Package          string
	ApiGroup         string
	IsCrd            bool
	NotAKubeResource bool
}

func (p Parameter) String() string {
	return string(p.LowerName)
}

// registered parameters
var Parameters []Parameter

func Register(param Parameter) Parameter {
	for _, existing := range Parameters {
		if existing.LowerName == param.LowerName {
			logrus.Fatalf("parameter %v already defined", param.LowerName)
		}
	}
	Parameters = append(Parameters, param)
	return param
}

var (
	ConfigMaps = Register(Parameter{
		LowerName:    "configmaps",
		PluralName:   "ConfigMaps",
		SingleName:   "ConfigMap",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	Services = Register(Parameter{
		LowerName:    "services",
		PluralName:   "Services",
		SingleName:   "Service",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	Pods = Register(Parameter{
		LowerName:    "pods",
		PluralName:   "Pods",
		SingleName:   "Pod",
		ImportPrefix: "corev1",
		Package:      "k8s.io/api/core/v1",
		ApiGroup:     "",
	})
	Deployments = Register(Parameter{
		LowerName:    "deployments",
		PluralName:   "Deployments",
		SingleName:   "Deployment",
		ImportPrefix: "appsv1",
		Package:      "k8s.io/api/apps/v1",
		ApiGroup:     "apps",
	})
	ReplicaSets = Register(Parameter{
		LowerName:    "replicasets",
		PluralName:   "ReplicaSets",
		SingleName:   "ReplicaSet",
		ImportPrefix: "appsv1",
		Package:      "k8s.io/api/apps/v1",
		ApiGroup:     "apps",
	})
	TrafficSplits = Register(Parameter{
		LowerName:    "trafficsplits",
		PluralName:   "TrafficSplits",
		SingleName:   "TrafficSplit",
		ImportPrefix: "trafficsplitv1alpha2",
		Package:      "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha2",
		ApiGroup:     "split.smi-spec.io",
		IsCrd:        true,
	})
	VirtualServices = Register(Parameter{
		LowerName:    "virtualservices",
		PluralName:   "VirtualServices",
		SingleName:   "VirtualService",
		ImportPrefix: "istiov1alpha3",
		Package:      "istio.io/client-go/pkg/apis/networking/v1alpha3",
		ApiGroup:     "networking.istio.io",
		IsCrd:        true,
	})
	Gateways = Register(Parameter{
		LowerName:    "gateways",
		PluralName:   "Gateways",
		SingleName:   "Gateway",
		ImportPrefix: "istiov1alpha3",
		Package:      "istio.io/client-go/pkg/apis/networking/v1alpha3",
		ApiGroup:     "networking.istio.io",
		IsCrd:        true,
	})
	Metrics = Register(Parameter{
		LowerName:        "metrics",
		PluralName:       "Metrics",
		SingleName:       "Metric",
		ImportPrefix:     "metrics",
		Package:          "github.com/solo-io/autopilot/pkg/metrics",
		NotAKubeResource: true,
	})
)
