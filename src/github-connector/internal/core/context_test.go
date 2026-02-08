package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSourceContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreateSourceContext()
	want := &Context{
		Id:           "github:source",
		Name:         "github:source",
		Level:        1,
		ParentId:     "",
		ConnectorId:  "github",
		ResourceType: "source",
		Title:        ptrString("GitHub"),
		Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
		Url:          ptrString("https://github.com"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateRepositoryContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreateRepositoryContext("octocat/Hello-World")
	want := &Context{
		Id:           "github:repository:octocat/Hello-World",
		Name:         "repository:octocat/Hello-World",
		Level:        2,
		ParentId:     "github:source",
		ConnectorId:  "github",
		ResourceType: "repository",
		Title:        ptrString("octocat/Hello-World"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"repo": "octocat/Hello-World",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreatePRContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreatePRContext("octocat/Hello-World", 42)
	want := &Context{
		Id:           "github:pull_request:octocat/Hello-World:42",
		Name:         "PR #42",
		Level:        3,
		ParentId:     "github:repository:octocat/Hello-World",
		ConnectorId:  "github",
		ResourceType: "pull_request",
		Title:        ptrString("PR #42"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"repo":      "octocat/Hello-World",
				"pr_number": "42",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateIssueContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreateIssueContext("octocat/Hello-World", 101)
	want := &Context{
		Id:           "github:issue:octocat/Hello-World:101",
		Name:         "Issue #101",
		Level:        3,
		ParentId:     "github:repository:octocat/Hello-World",
		ConnectorId:  "github",
		ResourceType: "issue",
		Title:        ptrString("Issue #101"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"repo":         "octocat/Hello-World",
				"issue_number": "101",
			},
		},
	}
	assert.Equal(t, want, got)
}
