package config

// EDIT THIS FILE!  THIS IS CONFIGURATION FOR YOU TO OWN!

import "time"

var (
	// the version of the controller
	Version = "0.0.1"

	// The Address of the Prometheus server to connect to for metrics
	LocalMetricsServer = "http://prometheus.istio-system:9090"

	// The Mesh Provider (determines what types of metrics to use)
	MeshProvider = "istio"

	// Modify the WorkInterval to change the interval at which workers resync
	WorkInterval = time.Second * 5
)
