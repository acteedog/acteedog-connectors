package match

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient is an in-test implementation of HTTPClient.
type mockHTTPClient struct {
	response *IssueResponse
	err      error
}

func (m *mockHTTPClient) FetchIssue(_, _, _, _ string) (*IssueResponse, error) {
	return m.response, m.err
}

func issueWithProject(id, key, projectID, projectName, issueTypeName string) *IssueResponse {
	return &IssueResponse{
		ID:  id,
		Key: key,
		Fields: IssueFields{
			Project:   &ProjectRef{ID: projectID, Key: "PROJ", Name: projectName},
			IssueType: &NamedRef{Name: issueTypeName},
		},
	}
}

func issueWithParent(id, key, projectID, projectName, issueTypeName, parentID, parentKey, parentTypeName string) *IssueResponse {
	return &IssueResponse{
		ID:  id,
		Key: key,
		Fields: IssueFields{
			Project:   &ProjectRef{ID: projectID, Key: "PROJ", Name: projectName},
			IssueType: &NamedRef{Name: issueTypeName},
			Parent: &IssueRef{
				ID:  parentID,
				Key: parentKey,
				Fields: &IssueFields{
					IssueType: &NamedRef{Name: parentTypeName},
				},
			},
		},
	}
}

func newMatcher(httpClient HTTPClient) *ContextMatcher {
	return NewContextMatcher(httpClient, "test-cloud-id", "myorg")
}

func TestMatchURL_NoMatch(t *testing.T) {
	matcher := newMatcher(&mockHTTPClient{})

	urls := []string{
		"https://example.com/browse/PROJ-1",
		"https://myorg.atlassian.net/jira/software/PROJ-1",
		"https://github.com/owner/repo/issues/1",
		"not-a-url",
	}

	for _, url := range urls {
		got, err := matcher.MatchURLWithCredentials(url, "", "")
		assert.NoError(t, err)
		assert.Empty(t, got, "should not match URL: %s", url)
	}
}

func TestMatchURL_DifferentSubdomain(t *testing.T) {
	matcher := newMatcher(&mockHTTPClient{})
	// URL belongs to a different Jira instance
	got, err := matcher.MatchURLWithCredentials("https://other-org.atlassian.net/browse/PROJ-1", "", "")
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestMatchURL_IssueWithProject(t *testing.T) {
	issue := issueWithProject("10001", "PROJ-1", "10000", "My Project", "Task")
	matcher := newMatcher(&mockHTTPClient{response: issue})

	got, err := matcher.MatchURLWithCredentials("https://myorg.atlassian.net/browse/PROJ-1", "", "")
	assert.NoError(t, err)
	assert.Len(t, got, 3)

	assert.Equal(t, "jira:source", got[0].Id)
	assert.Equal(t, "source", got[0].ResourceType)

	assert.Equal(t, "jira:project:10000", got[1].Id)
	assert.Equal(t, "project", got[1].ResourceType)
	assert.Equal(t, "jira:source", got[1].ParentId)

	assert.Equal(t, "jira:issue:10001", got[2].Id)
	assert.Equal(t, "issue", got[2].ResourceType)
	assert.Equal(t, "jira:project:10000", got[2].ParentId)
}

func TestMatchURL_IssueWithParent(t *testing.T) {
	issue := issueWithParent("10002", "PROJ-2", "10000", "My Project", "Task", "10001", "PROJ-1", "Epic")
	matcher := newMatcher(&mockHTTPClient{response: issue})

	got, err := matcher.MatchURLWithCredentials("https://myorg.atlassian.net/browse/PROJ-2", "", "")
	assert.NoError(t, err)
	assert.Len(t, got, 4)

	assert.Equal(t, "jira:source", got[0].Id)
	assert.Equal(t, "jira:project:10000", got[1].Id)
	// parent issue
	assert.Equal(t, "jira:issue:10001", got[2].Id)
	assert.Equal(t, "jira:project:10000", got[2].ParentId)
	// child issue
	assert.Equal(t, "jira:issue:10002", got[3].Id)
	assert.Equal(t, "jira:issue:10001", got[3].ParentId)
}

func TestMatchURL_EnrichmentParams(t *testing.T) {
	issue := issueWithProject("10001", "PROJ-1", "10000", "My Project", "Task")
	matcher := newMatcher(&mockHTTPClient{response: issue})

	got, err := matcher.MatchURLWithCredentials("https://myorg.atlassian.net/browse/PROJ-1", "", "")
	assert.NoError(t, err)
	assert.Len(t, got, 3)

	// Project enrichment_params
	projectMeta, ok := got[1].Metadata.(map[string]any)
	assert.True(t, ok)
	projectParams, ok := projectMeta["enrichment_params"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "10000", projectParams["project_id"])

	// Issue enrichment_params
	issueMeta, ok := got[2].Metadata.(map[string]any)
	assert.True(t, ok)
	issueParams, ok := issueMeta["enrichment_params"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "10001", issueParams["issue_id"])
}

func TestMatchURL_APIError(t *testing.T) {
	matcher := newMatcher(&mockHTTPClient{err: fmt.Errorf("API unreachable")})

	got, err := matcher.MatchURLWithCredentials("https://myorg.atlassian.net/browse/PROJ-1", "", "")
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestMatchURL_MissingProject(t *testing.T) {
	// Issue with no project info: should return only source context
	issue := &IssueResponse{ID: "10001", Key: "PROJ-1"}
	matcher := newMatcher(&mockHTTPClient{response: issue})

	got, err := matcher.MatchURLWithCredentials("https://myorg.atlassian.net/browse/PROJ-1", "", "")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "jira:source", got[0].Id)
}

func TestMatchURL_IssueKeyFormats(t *testing.T) {
	issue := issueWithProject("10001", "PROJ-1", "10000", "My Project", "Task")
	matcher := newMatcher(&mockHTTPClient{response: issue})

	validKeys := []string{
		"https://myorg.atlassian.net/browse/PROJ-1",
		"https://myorg.atlassian.net/browse/MYPROJ-123",
		"https://myorg.atlassian.net/browse/AB-1",
	}
	for _, url := range validKeys {
		got, err := matcher.MatchURLWithCredentials(url, "", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, got, "should match: %s", url)
	}

	invalidKeys := []string{
		"https://myorg.atlassian.net/browse/proj-1", // lowercase project key
		"https://myorg.atlassian.net/browse/PROJ",   // no issue number
		"https://myorg.atlassian.net/browse/1-PROJ", // number first
	}
	for _, url := range invalidKeys {
		got, err := matcher.MatchURLWithCredentials(url, "", "")
		assert.NoError(t, err)
		assert.Empty(t, got, "should not match: %s", url)
	}
}
