package fetch

type HTTPClient interface {
	FetchActivities(token, username string, page int) ([]map[string]any, error)
}
