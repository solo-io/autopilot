package metrics

import (
	v1 "github.com/solo-io/autopilot/api/v1"
	"os"
	"time"
)

type Factory struct {
	MeshProvider v1.MeshProvider
	Client       *PrometheusClient
}

func getMetricsServer(meshProvider v1.MeshProvider) string {
	if metricsServer := os.Getenv("METRICS_SERVER"); metricsServer != "" {
		return metricsServer
	}
	switch meshProvider {
	case v1.MeshProvider_Istio:
		return "http://prometheus.istio-system:9090"
	}
	panic("currently unsupported: " + meshProvider.String())
}

func NewFactory(meshProvider v1.MeshProvider, timeout time.Duration) (*Factory, error) {
	metricsAddr := getMetricsServer(meshProvider)
	client, err := NewPrometheusClient(metricsAddr, timeout)
	if err != nil {
		return nil, err
	}

	return &Factory{
		MeshProvider: meshProvider,
		Client:       client,
	}, nil
}

func (factory *Factory) Observer() Metrics {
	provider := factory.MeshProvider
	switch provider {
	case v1.MeshProvider_Istio:
		return &IstioObserver{
			client: factory.Client,
		}
	default:
		return &HttpObserver{
			client: factory.Client,
		}
	}
}
