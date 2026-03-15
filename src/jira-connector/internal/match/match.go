package match

import (
	"fmt"
	"jira-connector/internal/core"
	"regexp"
	"strings"
)

// Jira browse URL pattern: https://<subdomain>.atlassian.net/browse/<ISSUE-KEY>
var reBrowseURL = regexp.MustCompile(`^https://([^.]+)\.atlassian\.net/browse/([A-Z][A-Z0-9_]+-\d+)`)

// ContextMatcher handles URL matching and context building for Jira.
type ContextMatcher struct {
	httpClient    HTTPClient
	cloudID       string
	siteSubdomain string
}

// NewContextMatcher creates a new ContextMatcher.
func NewContextMatcher(httpClient HTTPClient, cloudID, siteSubdomain string) *ContextMatcher {
	return &ContextMatcher{
		httpClient:    httpClient,
		cloudID:       cloudID,
		siteSubdomain: siteSubdomain,
	}
}

// MatchURL returns context nodes for a single URL if it belongs to this Jira instance.
// Returns an empty slice when the URL does not match.
func (m *ContextMatcher) MatchURL(url string) ([]*core.Context, error) {
	captures := reBrowseURL.FindStringSubmatch(url)
	if captures == nil {
		return []*core.Context{}, nil
	}

	subdomain := captures[1]
	issueKey := captures[2]

	// Only process URLs belonging to the configured site
	if !strings.EqualFold(subdomain, m.siteSubdomain) {
		return []*core.Context{}, nil
	}

	issue, err := m.httpClient.FetchIssue(m.cloudID, "", "", issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue %s: %w", issueKey, err)
	}

	return m.buildContexts(issue), nil
}

// MatchURLWithCredentials is like MatchURL but passes credentials to the HTTP client.
func (m *ContextMatcher) MatchURLWithCredentials(url, email, apiToken string) ([]*core.Context, error) {
	captures := reBrowseURL.FindStringSubmatch(url)
	if captures == nil {
		return []*core.Context{}, nil
	}

	subdomain := captures[1]
	issueKey := captures[2]

	if !strings.EqualFold(subdomain, m.siteSubdomain) {
		return []*core.Context{}, nil
	}

	issue, err := m.httpClient.FetchIssue(m.cloudID, email, apiToken, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue %s: %w", issueKey, err)
	}

	return m.buildContexts(issue), nil
}

// buildContexts constructs the context hierarchy (source > project > [parent issue >] issue).
func (m *ContextMatcher) buildContexts(issue *IssueResponse) []*core.Context {
	gen := core.NewContextGenerator(m.cloudID)
	source := gen.CreateSourceContext()

	if issue.Fields.Project == nil {
		return []*core.Context{source}
	}

	projectCtx := gen.CreateProjectContext(issue.Fields.Project.ID, issue.Fields.Project.Name)

	issueTypeName := "Issue"
	if issue.Fields.IssueType != nil {
		issueTypeName = issue.Fields.IssueType.Name
	}

	// If the issue has a parent issue, include that parent in the hierarchy
	if issue.Fields.Parent != nil {
		parent := issue.Fields.Parent
		parentTypeName := "Issue"
		if parent.Fields != nil && parent.Fields.IssueType != nil {
			parentTypeName = parent.Fields.IssueType.Name
		}
		parentCtx := gen.CreateParentIssueContext(parent.ID, parent.Key, parent.Key, issue.Fields.Project.ID, parentTypeName)
		issueCtx := gen.CreateIssueContextWithIssueParent(issue.ID, issue.Key, issue.Key, parent.ID, issueTypeName)
		return []*core.Context{source, projectCtx, parentCtx, issueCtx}
	}

	issueCtx := gen.CreateIssueContextWithProjectParent(issue.ID, issue.Key, issue.Key, issue.Fields.Project.ID, issueTypeName)
	return []*core.Context{source, projectCtx, issueCtx}
}
