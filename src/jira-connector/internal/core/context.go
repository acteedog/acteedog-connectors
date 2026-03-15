package core

import (
	"fmt"
	"time"
)

// Context represents a context object
type Context struct {
	ConnectorId  string
	CreatedAt    *time.Time
	Description  *string
	Id           string
	Level        int64
	Metadata     any
	Name         string
	ParentId     string
	ResourceType string
	Title        *string
	UpdatedAt    *time.Time
	Url          *string
}

// ContextGenerator provides factory methods for creating standardized Context objects
type ContextGenerator struct {
	connectorID string
	cloudID     string
}

// NewContextGenerator creates a new ContextGenerator
func NewContextGenerator(cloudID string) *ContextGenerator {
	return &ContextGenerator{
		connectorID: ConnectorID,
		cloudID:     cloudID,
	}
}

// CreateSourceContext creates a Level 1 source context for Jira
func (g *ContextGenerator) CreateSourceContext() *Context {
	id := MakeSourceContextID()
	url := fmt.Sprintf("%s/%s", JiraAPIBase, g.cloudID)
	title := "Jira"
	description := "Activity source from Jira"
	return &Context{
		Id:           id,
		Name:         id,
		Level:        1,
		ParentId:     "",
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeSource,
		Title:        &title,
		Description:  &description,
		Url:          &url,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
}

// CreateProjectContext creates a Level 2 project context
func (g *ContextGenerator) CreateProjectContext(projectID, projectName string) *Context {
	id := MakeProjectContextID(projectID)
	parentID := MakeSourceContextID()
	name := fmt.Sprintf("project %s", projectName)
	title := projectName
	return &Context{
		Id:           id,
		Name:         name,
		Level:        2,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeProject,
		Title:        &title,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"project_id": projectID,
			},
		},
	}
}

// issueContextName generates a context name for an issue based on its type and key
func issueContextName(issueTypeName, issueKey string) string {
	return fmt.Sprintf("%s %s", issueTypeName, issueKey)
}

// CreateIssueContextWithProjectParent creates a Level 3 issue context with a project as parent
func (g *ContextGenerator) CreateIssueContextWithProjectParent(issueID, issueKey, summary, projectID, issueTypeName string) *Context {
	id := MakeIssueContextID(issueID)
	parentID := MakeProjectContextID(projectID)
	name := issueContextName(issueTypeName, issueKey)
	title := summary
	return &Context{
		Id:           id,
		Name:         name,
		Level:        3,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeIssue,
		Title:        &title,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"issue_id": issueID,
			},
		},
	}
}

// CreateIssueContextWithIssueParent creates a Level 3 issue context with a parent issue as parent
func (g *ContextGenerator) CreateIssueContextWithIssueParent(issueID, issueKey, summary, parentIssueID, issueTypeName string) *Context {
	id := MakeIssueContextID(issueID)
	parentID := MakeIssueContextID(parentIssueID)
	name := issueContextName(issueTypeName, issueKey)
	title := summary
	return &Context{
		Id:           id,
		Name:         name,
		Level:        3,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeIssue,
		Title:        &title,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"issue_id": issueID,
			},
		},
	}
}

// CreateParentIssueContext creates a Level 3 parent issue context with a project as parent
func (g *ContextGenerator) CreateParentIssueContext(issueID, issueKey, summary, projectID, issueTypeName string) *Context {
	id := MakeIssueContextID(issueID)
	parentID := MakeProjectContextID(projectID)
	name := issueContextName(issueTypeName, issueKey)
	title := summary
	return &Context{
		Id:           id,
		Name:         name,
		Level:        3,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeIssue,
		Title:        &title,
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"issue_id": issueID,
			},
		},
	}
}
