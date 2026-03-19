package enrich

// HTTPClient is the interface for fetching GitHub resource data for enrichment.
type HTTPClient interface {
	FetchRepository(repo string) (map[string]any, error)
	FetchPullRequest(repo, number string) (map[string]any, error)
	FetchIssue(repo, number string) (map[string]any, error)
}
