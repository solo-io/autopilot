// +build ignore

package metrics

import "github.com/prometheus/common/model"

// the result of a generic metrics query
// a shim for any value that can be returned by PromQL
type QueryResult struct {
	// value returned is a prometheus query result
	model.Value
}
