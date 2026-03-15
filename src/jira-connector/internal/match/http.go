package match

// HTTPClient defines the interface for Jira API calls needed during context matching.
type HTTPClient interface {
	FetchIssue(cloudID, email, apiToken, issueKey string) (*IssueResponse, error)
}

// IssueResponse is a minimal Jira issue API response for context matching.
type IssueResponse struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

// IssueFields contains the fields needed to build a context hierarchy.
type IssueFields struct {
	Project   *ProjectRef `json:"project"`
	Parent    *IssueRef   `json:"parent"`
	IssueType *NamedRef   `json:"issuetype"`
}

// ProjectRef represents a project reference in an issue's fields.
type ProjectRef struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// IssueRef represents a parent issue reference in an issue's fields.
type IssueRef struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Fields *IssueFields `json:"fields"`
}

// NamedRef represents a named Jira resource (e.g. issue type).
type NamedRef struct {
	Name string `json:"name"`
}
