package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptrString(s string) *string {
	return &s
}

func TestCreateSourceContext(t *testing.T) {
	g := NewContextGenerator("my-cloud-id")
	got := g.CreateSourceContext()
	want := &Context{
		Id:           "jira:source",
		Name:         "jira:source",
		Level:        1,
		ParentId:     "",
		ConnectorId:  "jira",
		ResourceType: "source",
		Title:        ptrString("Jira"),
		Description:  ptrString("Activity source from Jira"),
		Url:          ptrString("https://api.atlassian.com/ex/jira/my-cloud-id"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateProjectContext(t *testing.T) {
	g := NewContextGenerator("my-cloud-id")
	got := g.CreateProjectContext("10000", "test-project")
	want := &Context{
		Id:           "jira:project:10000",
		Name:         "project test-project",
		Level:        2,
		ParentId:     "jira:source",
		ConnectorId:  "jira",
		ResourceType: "project",
		Title:        ptrString("test-project"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"project_id": "10000",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateIssueContextWithProjectParent(t *testing.T) {
	g := NewContextGenerator("my-cloud-id")
	got := g.CreateIssueContextWithProjectParent("10038", "TES-6", "Test Epic", "10000", "Epic")
	want := &Context{
		Id:           "jira:issue:10038",
		Name:         "Epic TES-6",
		Level:        3,
		ParentId:     "jira:project:10000",
		ConnectorId:  "jira",
		ResourceType: "issue",
		Title:        ptrString("Test Epic"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"issue_id": "10038",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateIssueContextWithIssueParent(t *testing.T) {
	g := NewContextGenerator("my-cloud-id")
	got := g.CreateIssueContextWithIssueParent("10003", "TES-1", "Sub-task", "10000", "Subtask")
	want := &Context{
		Id:           "jira:issue:10003",
		Name:         "Subtask TES-1",
		Level:        3,
		ParentId:     "jira:issue:10000",
		ConnectorId:  "jira",
		ResourceType: "issue",
		Title:        ptrString("Sub-task"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"issue_id": "10003",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateParentIssueContext(t *testing.T) {
	g := NewContextGenerator("my-cloud-id")
	got := g.CreateParentIssueContext("10000", "TES-5", "Parent Epic", "20000", "Epic")
	want := &Context{
		Id:           "jira:issue:10000",
		Name:         "Epic TES-5",
		Level:        3,
		ParentId:     "jira:project:20000",
		ConnectorId:  "jira",
		ResourceType: "issue",
		Title:        ptrString("Parent Epic"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"issue_id": "10000",
			},
		},
	}
	assert.Equal(t, want, got)
}
