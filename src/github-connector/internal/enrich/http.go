package enrich

type HTTPClient interface {
	FetchRepository(token, repo string) (map[string]any, error)
	FetchPullRequest(token, repo, number string) (map[string]any, error)
	FetchIssue(token, repo, number string) (map[string]any, error)
}
