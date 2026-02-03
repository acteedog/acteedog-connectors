package core

import (
	"fmt"
)

const (
	// ConnectorVersion is the version of this connector
	ConnectorVersion = "0.1.0"
	// ConnectorID is the unique identifier for this connector
	ConnectorID = "github"
	// GithubAPIBaseURL is the base URL for GitHub API
	GithubAPIBaseURL = "https://api.github.com"
)

// Resource type constants for context identification
const (
	ResourceTypeSource      = "source"
	ResourceTypeRepository  = "repository"
	ResourceTypePullRequest = "pull_request"
	ResourceTypeIssue       = "issue"
)

// MakeActivityID creates an activity ID with connector prefix
func MakeActivityID(eventID string) string {
	return fmt.Sprintf("%s:%s", ConnectorID, eventID)
}

// MakeSourceContextID creates a source context ID with connector prefix
func MakeSourceContextID() string {
	return fmt.Sprintf("%s:%s", ConnectorID, ResourceTypeSource)
}

// MakeRepositoryContextID creates a repository context ID with connector prefix
func MakeRepositoryContextID(repoName string) string {
	return fmt.Sprintf("%s:%s:%s", ConnectorID, ResourceTypeRepository, repoName)
}

// MakePullRequestContextID creates a pull request context ID with connector prefix
func MakePullRequestContextID(repoName, prNumber string) string {
	return fmt.Sprintf("%s:%s:%s:%s", ConnectorID, ResourceTypePullRequest, repoName, prNumber)
}

// MakeIssueContextID creates an issue context ID with connector prefix
func MakeIssueContextID(repoName, issueNumber string) string {
	return fmt.Sprintf("%s:%s:%s:%s", ConnectorID, ResourceTypeIssue, repoName, issueNumber)
}
