package initializing

// TODO generate

import (
	"github.com/solo-io/autopilot/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	trafficsplitv1alpha2 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha2"
)

type Inputs struct {
	Metrics metrics.Metrics
}

type Outputs struct {
	Deployments   appsv1.DeploymentList
	Services      corev1.ServiceList
	TrafficSplits trafficsplitv1alpha2.TrafficSplitList
}
