package main

// MatchContext matches the provided URLs for the Slack connector.
// Currently not supported: all URLs return empty context lists.
func MatchContext(input MatchContextRequest) (MatchContextResponse, error) {
	results := make([]MatchContextResult, 0, len(input.Urls))
	for _, url := range input.Urls {
		results = append(results, MatchContextResult{
			Url:      url,
			Contexts: []Context{},
		})
	}
	return MatchContextResponse{Results: results}, nil
}
