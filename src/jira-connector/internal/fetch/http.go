package fetch

type HTTPClient interface {
	FetchIssues(cloudID, email, apiToken string, projectIDs []string, dateFrom, dateTo string) (*JiraSearchResponse, error)
}
