package enrich

type HTTPClient interface {
	FetchProject(cloudID, email, apiToken, projectID string) (*JiraProjectResponse, error)
	FetchIssue(cloudID, email, apiToken, issueID string) (*JiraIssueResponse, error)
}
