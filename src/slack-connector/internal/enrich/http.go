package enrich

type HTTPClient interface {
	FetchChannel(token, channelID string) (map[string]any, error)
	FetchThread(token, channelID, threadTS string) (map[string]any, error)
}
