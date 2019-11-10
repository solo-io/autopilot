package metrics

// A generic interface for interacting with Metrics stores.
type Client interface {
	RunQuery(queryTemplate string, data map[string]string) (*QueryResult, error)
}

func runQuery(queryTemplate string, data map[string]string) (*QueryResult, error) {

}
