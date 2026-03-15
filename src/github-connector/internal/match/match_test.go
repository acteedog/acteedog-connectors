package match

import (
	"github-connector/internal/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptrString(s string) *string { return &s }

func gen() *core.ContextGenerator { return core.NewContextGenerator() }

// --- Pull Request ---

func TestMatchURL_PullRequest_Basic(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/pull/42")
	assert.Equal(t, []*core.Context{
		{
			Id:           "github:source",
			Name:         "github:source",
			ParentId:     "",
			ConnectorId:  "github",
			ResourceType: "source",
			Title:        ptrString("GitHub"),
			Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
			Url:          ptrString("https://github.com"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{}},
		},
		{
			Id:           "github:repository:octocat/Hello-World",
			Name:         "repository:octocat/Hello-World",
			ParentId:     "github:source",
			ConnectorId:  "github",
			ResourceType: "repository",
			Title:        ptrString("octocat/Hello-World"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{"repo": "octocat/Hello-World"}},
		},
		{
			Id:           "github:pull_request:octocat/Hello-World:42",
			Name:         "PR #42",
			ParentId:     "github:repository:octocat/Hello-World",
			ConnectorId:  "github",
			ResourceType: "pull_request",
			Title:        ptrString("PR #42"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{"repo": "octocat/Hello-World", "pr_number": "42"}},
		},
	}, got)
}

func TestMatchURL_PullRequest_TrailingSlash(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/pull/42/")
	assert.Len(t, got, 3)
	assert.Equal(t, "github:pull_request:octocat/Hello-World:42", got[2].Id)
}

func TestMatchURL_PullRequest_SlackMrkdwn(t *testing.T) {
	got := MatchURL(gen(), "<https://github.com/octocat/Hello-World/pull/42|PR #42>")
	assert.Len(t, got, 3)
	assert.Equal(t, "github:pull_request:octocat/Hello-World:42", got[2].Id)
}

func TestMatchURL_PullRequest_NotMatchIssueURL(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/issues/42")
	// Issue URL should not produce a pull_request context
	if assert.Len(t, got, 3) {
		assert.Equal(t, "issue", got[2].ResourceType)
	}
}

// --- Issue ---

func TestMatchURL_Issue_Basic(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/issues/42")
	assert.Equal(t, []*core.Context{
		{
			Id:           "github:source",
			Name:         "github:source",
			ParentId:     "",
			ConnectorId:  "github",
			ResourceType: "source",
			Title:        ptrString("GitHub"),
			Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
			Url:          ptrString("https://github.com"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{}},
		},
		{
			Id:           "github:repository:octocat/Hello-World",
			Name:         "repository:octocat/Hello-World",
			ParentId:     "github:source",
			ConnectorId:  "github",
			ResourceType: "repository",
			Title:        ptrString("octocat/Hello-World"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{"repo": "octocat/Hello-World"}},
		},
		{
			Id:           "github:issue:octocat/Hello-World:42",
			Name:         "Issue #42",
			ParentId:     "github:repository:octocat/Hello-World",
			ConnectorId:  "github",
			ResourceType: "issue",
			Title:        ptrString("Issue #42"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{"repo": "octocat/Hello-World", "issue_number": "42"}},
		},
	}, got)
}

func TestMatchURL_Issue_TrailingSlash(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/issues/42/")
	assert.Len(t, got, 3)
	assert.Equal(t, "github:issue:octocat/Hello-World:42", got[2].Id)
}

func TestMatchURL_Issue_SlackMrkdwn(t *testing.T) {
	got := MatchURL(gen(), "<https://github.com/octocat/Hello-World/issues/42|Issue #42>")
	assert.Len(t, got, 3)
	assert.Equal(t, "github:issue:octocat/Hello-World:42", got[2].Id)
}

func TestMatchURL_Issue_NotMatchPullRequestURL(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/pull/42")
	if assert.Len(t, got, 3) {
		assert.Equal(t, "pull_request", got[2].ResourceType)
	}
}

// --- Repository ---

func TestMatchURL_Repository_Basic(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World")
	assert.Equal(t, []*core.Context{
		{
			Id:           "github:source",
			Name:         "github:source",
			ParentId:     "",
			ConnectorId:  "github",
			ResourceType: "source",
			Title:        ptrString("GitHub"),
			Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
			Url:          ptrString("https://github.com"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{}},
		},
		{
			Id:           "github:repository:octocat/Hello-World",
			Name:         "repository:octocat/Hello-World",
			ParentId:     "github:source",
			ConnectorId:  "github",
			ResourceType: "repository",
			Title:        ptrString("octocat/Hello-World"),
			Metadata:     map[string]any{"enrichment_params": map[string]any{"repo": "octocat/Hello-World"}},
		},
	}, got)
}

func TestMatchURL_Repository_TrailingSlash(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

func TestMatchURL_Repository_SlackMrkdwn(t *testing.T) {
	got := MatchURL(gen(), "<https://github.com/octocat/Hello-World|My Repository>")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

func TestMatchURL_Repository_EndWithCloseParen(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World)")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

func TestMatchURL_Repository_EndWithCloseBracket(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World]")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

func TestMatchURL_Repository_EndWithDoubleQuote(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World\"")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

func TestMatchURL_Repository_EndWithSingleQuote(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World'")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

func TestMatchURL_Repository_EndWithQueryParam(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World?param=value")
	assert.Len(t, got, 2)
	assert.Equal(t, "github:repository:octocat/Hello-World", got[1].Id)
}

// --- Exclude pattern ---

func TestMatchURL_ExcludePattern_UserAttachmentsAsset(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/user-attachments/assets/a548f44f-5f9d-4ad0-8c2b-e7211b7bc08b")
	assert.Empty(t, got)
}

func TestMatchURL_ExcludePattern_UserAttachmentsNoPath(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/user-attachments/")
	assert.Empty(t, got)
}

func TestMatchURL_ExcludePattern_RegularRepoNotExcluded(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World")
	assert.NotEmpty(t, got)
}

func TestMatchURL_ExcludePattern_PullRequestNotExcluded(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/pull/42")
	assert.NotEmpty(t, got)
}

// --- No match ---

func TestMatchURL_NoMatch(t *testing.T) {
	urls := []string{
		"https://example.com/foo/bar",
		"https://gitlab.com/owner/repo/pull/1",
		"not-a-url",
		"https://github.com",
	}
	for _, url := range urls {
		got := MatchURL(gen(), url)
		assert.Empty(t, got, "should not match: %s", url)
	}
}

// --- Context hierarchy integrity ---

func TestMatchURL_ContextHierarchy_PR(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/pull/42")
	assert.Len(t, got, 3)
	// source has no parent
	assert.Equal(t, "", got[0].ParentId)
	// repository's parent is source
	assert.Equal(t, got[0].Id, got[1].ParentId)
	// PR's parent is repository
	assert.Equal(t, got[1].Id, got[2].ParentId)
}

func TestMatchURL_ContextHierarchy_Issue(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World/issues/101")
	assert.Len(t, got, 3)
	assert.Equal(t, "", got[0].ParentId)
	assert.Equal(t, got[0].Id, got[1].ParentId)
	assert.Equal(t, got[1].Id, got[2].ParentId)
}

func TestMatchURL_ContextHierarchy_Repository(t *testing.T) {
	got := MatchURL(gen(), "https://github.com/octocat/Hello-World")
	assert.Len(t, got, 2)
	assert.Equal(t, "", got[0].ParentId)
	assert.Equal(t, got[0].Id, got[1].ParentId)
}
