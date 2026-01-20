package main

// GetContextPatterns returns URL patterns for Slack (empty for now)
// This function is required by the XTP schema but currently not supported in Slack connector.
func GetContextPatterns() (ContextPatternsResponse, error) {
	return ContextPatternsResponse{
		Patterns: []ContextPatternDefinition{},
	}, nil
}
