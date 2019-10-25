package metrics

import "time"

var glooQueries = map[string]string{
	"request-success-rate": `
	sum(
		rate(
			envoy_cluster_upstream_rq{
				envoy_cluster_name=~"{{ .Namespace }}-{{ .Name }}-canary-[0-9a-zA-Z-]+_[0-9a-zA-Z-]+",
				envoy_response_code!~"5.*"
			}[{{ .Interval }}]
		)
	) 
	/ 
	sum(
		rate(
			envoy_cluster_upstream_rq{
				envoy_cluster_name=~"{{ .Namespace }}-{{ .Name }}-canary-[0-9a-zA-Z-]+_[0-9a-zA-Z-]+",
			}[{{ .Interval }}]
		)
	) 
	* 100`,
	"request-duration": `
	histogram_quantile(
		0.99,
		sum(
			rate(
				envoy_cluster_upstream_rq_time_bucket{
					envoy_cluster_name=~"{{ .Namespace }}-{{ .Name }}-canary-[0-9a-zA-Z-]+_[0-9a-zA-Z-]+",
				}[{{ .Interval }}]
			)
		) by (le)
	)`,
}

type GlooObserver struct {
	Client *PrometheusClient
}

func NewGlooObserver(client *PrometheusClient) *GlooObserver {
	return &GlooObserver{Client: client}
}

func (ob *GlooObserver) GetRequestSuccessRate(name string, namespace string, interval string) (float64, error) {
	query, err := ob.Client.RenderQuery(name, namespace, interval, glooQueries["request-success-rate"])
	if err != nil {
		return 0, err
	}

	value, err := ob.Client.RunQuery(query)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (ob *GlooObserver) GetRequestDuration(name string, namespace string, interval string) (time.Duration, error) {
	query, err := ob.Client.RenderQuery(name, namespace, interval, glooQueries["request-duration"])
	if err != nil {
		return 0, err
	}

	value, err := ob.Client.RunQuery(query)
	if err != nil {
		return 0, err
	}

	ms := time.Duration(int64(value)) * time.Millisecond
	return ms, nil
}
