package main

// GetContextPatterns returns URL patterns for Jira (empty for now)
// This function is required by the XTP schema but context detection is done via activity-based approach.
func GetContextPatterns() (ContextPatternsResponse, error) {
	return ContextPatternsResponse{
		Patterns: []ContextPatternDefinition{},
	}, nil
}
