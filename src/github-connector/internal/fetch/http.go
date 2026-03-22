package fetch

// HTTPClient is the interface for fetching GitHub events.
type HTTPClient interface {
	FetchActivities(username string, page int) ([]map[string]any, error)
}
