package main

// MatchContext returns empty contexts for all URLs.
// Google Calendar does not have a URL-based context matching strategy.
func MatchContext(input MatchContextRequest) (MatchContextResponse, error) {
	results := make([]MatchContextResult, 0, len(input.Urls))
	for _, u := range input.Urls {
		results = append(results, MatchContextResult{
			Url:      u,
			Contexts: []Context{},
		})
	}
	return MatchContextResponse{Results: results}, nil
}
