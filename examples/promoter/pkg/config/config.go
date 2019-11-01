package config

import "time"

// The Address of the Prometheus server to connect to for metrics
var MetricsServer = "http://prometheus:9090"

// The Mesh Provider (determines what types of metrics to use)
var MeshProvider = "istio"

// Modify the WorkInterval to change the interval at which workers resync
var WorkInterval = time.Second * 5
