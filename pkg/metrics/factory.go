package metrics

import (
	"fmt"
	v1 "github.com/solo-io/autopilot/api/v1"
	"os"
	"time"
)

type Factory struct {
	MeshProvider v1.MeshProvider
	Client       *PrometheusClient
}

func getMetricsServer(meshProvider v1.MeshProvider, controlPlaneNs string) string {
	if metricsServer := os.Getenv("METRICS_SERVER"); metricsServer != "" {
		return metricsServer
	}
	switch meshProvider {
	case v1.MeshProvider_Istio:
		return fmt.Sprintf("http://prometheus.%v:9090", controlPlaneNs)
	}
	panic("currently unsupported: " + meshProvider.String())
}

func NewFactory(cfg *v1.AutoPilotOperator, timeout time.Duration) (*Factory, error) {
	metricsAddr := getMetricsServer(cfg.MeshProvider, cfg.ControlPlaneNs)
	client, err := NewPrometheusClient(metricsAddr, timeout)
	if err != nil {
		return nil, err
	}

	return &Factory{
		MeshProvider: cfg.MeshProvider,
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
