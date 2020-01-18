// +build ignore

package metrics

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// A generic interface for interacting with Metrics stores.
type Client interface {
	RunQuery(ctx context.Context, queryTemplate string, data map[string]string) (*QueryResult, error)
}

type promClient struct {
	v1.API
}

// returns a client for running queries against Prometheus
func NewPrometheusClient(addr string) (*promClient, error) {
	client, err := api.NewClient(api.Config{
		Address: addr,
	})
	if err != nil {
		return nil, err
	}
	return &promClient{API: v1.NewAPI(client)}, nil
}

func (c *promClient) RunQuery(ctx context.Context, queryTemplate string, data map[string]string) (*QueryResult, error) {
	query, err := renderQuery(queryTemplate, data)
	if err != nil {
		return nil, errors.Wrapf(err, "rendering query")
	}
	value, _, err := c.API.Query(ctx, query, time.Now())
	return &QueryResult{Value: value}, err
}

func renderQuery(queryTemplate string, data map[string]string) (string, error) {
	tmpl := template.Must(template.New("query").Parse(queryTemplate))
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	return buf.String(), err
}
