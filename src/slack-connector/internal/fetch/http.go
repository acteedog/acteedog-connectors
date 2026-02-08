package fetch

type HTTPClient interface {
	FetchMessages(token, userID, targetDate string, page int) (map[string]any, error)
}
