package core

import "fmt"

const (
	// ConnectorID is the unique identifier for this connector
	ConnectorID = "jira"
	// JiraAPIBase is the base URL for Atlassian Cloud Jira REST API
	JiraAPIBase = "https://api.atlassian.com/ex/jira"
)

// Resource type constants for context identification
const (
	ResourceTypeSource  = "source"
	ResourceTypeProject = "project"
	ResourceTypeIssue   = "issue"
)

// MakeActivityID creates an activity ID for issue creation
func MakeIssueCreatedActivityID(projectID, issueID string) string {
	return fmt.Sprintf("%s:project:%s:issue:%s:created", ConnectorID, projectID, issueID)
}

// MakeCommentActivityID creates an activity ID for a comment
func MakeCommentActivityID(projectID, issueID, commentID string) string {
	return fmt.Sprintf("%s:project:%s:issue:%s:comment:%s", ConnectorID, projectID, issueID, commentID)
}

// MakeStatusChangedActivityID creates an activity ID for a status change
func MakeStatusChangedActivityID(projectID, issueID, historyID string) string {
	return fmt.Sprintf("%s:project:%s:issue:%s:status_changed:%s", ConnectorID, projectID, issueID, historyID)
}

// MakeSourceContextID creates a source context ID
func MakeSourceContextID() string {
	return fmt.Sprintf("%s:%s", ConnectorID, ResourceTypeSource)
}

// MakeProjectContextID creates a project context ID
func MakeProjectContextID(projectID string) string {
	return fmt.Sprintf("%s:%s:%s", ConnectorID, ResourceTypeProject, projectID)
}

// MakeIssueContextID creates an issue context ID
func MakeIssueContextID(projectID, issueID string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", ConnectorID, ResourceTypeProject, projectID, ResourceTypeIssue, issueID)
}
