package model

import v1 "github.com/solo-io/autopilot/api/v1"

// Default queries are built-in to the system and will be generated for any user metrics client
var DefaultQueries = []v1.MetricsQuery{
	{
		Name: "istio-success-rate",
		QueryTemplate: `sum(
		rate(
			istio_requests_total{
				destination_workload_namespace="{{ .Namespace }}",
				destination_workload=~"{{ .Name }}",
				response_code!~"5.*"
			}[{{ .Interval }}]
		)
	) 
	/ 
	sum(
		rate(
			istio_requests_total{
				destination_workload_namespace="{{ .Namespace }}",
				destination_workload=~"{{ .Name }}"
			}[{{ .Interval }}]
		)
	) 
	* 100`,
		Parameters: []string{
			"Namespace",
			"Name",
			"Interval",
		},
	},
	{
		Name: "istio-request_duration",
		QueryTemplate: `histogram_quantile(
		0.99,
		sum(
			rate(
				istio_request_duration_seconds_bucket{
					destination_workload_namespace="{{ .Namespace }}",
					destination_workload=~"{{ .Name }}"
				}[{{ .Interval }}]
			)
		) by (le)
	)`,
		Parameters: []string{
			"Namespace",
			"Name",
			"Interval",
		},
	},
	{
		Name: "envoy-success-rate",
		QueryTemplate: `sum(
		rate(
			envoy_cluster_upstream_rq{
				kubernetes_namespace="{{ .Namespace }}",
				kubernetes_pod_name=~"{{ .Name }}-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)",
				envoy_response_code!~"5.*"
			}[{{ .Interval }}]
		)
	) 
	/ 
	sum(
		rate(
			envoy_cluster_upstream_rq{
				kubernetes_namespace="{{ .Namespace }}",
				kubernetes_pod_name=~"{{ .Name }}-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)"
			}[{{ .Interval }}]
		)
	) 
	* 100`,
		Parameters: []string{
			"Namespace",
			"Name",
			"Interval",
		},
	},
	{
		Name: "envoy-request-duration",
		QueryTemplate: `histogram_quantile(
		0.99,
		sum(
			rate(
				envoy_cluster_upstream_rq_time_bucket{
					kubernetes_namespace="{{ .Namespace }}",
					kubernetes_pod_name=~"{{ .Name }}-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)"
				}[{{ .Interval }}]
			)
		) by (le)
	)`,
		Parameters: []string{
			"Namespace",
			"Name",
			"Interval",
		},
	},
}
