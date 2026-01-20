package main

import (
	"fmt"
)

// ContextGenerator provides factory methods for creating standardized Context objects
type ContextGenerator struct {
	connectorID string
}

// NewContextGenerator creates a new ContextGenerator
func NewContextGenerator() *ContextGenerator {
	return &ContextGenerator{
		connectorID: ConnectorID,
	}
}

// CreateSourceContext creates a Level 1 source context for GitHub
func (g *ContextGenerator) CreateSourceContext() Context {
	id := makeContextID(ResourceTypeSource, "")
	return Context{
		Id:           id,
		Name:         id,
		Level:        1,
		ParentId:     "", // Top level - no parent
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeSource,
		Title:        ptrString("GitHub"),
		Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
		Url:          ptrString("https://github.com"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
}

// CreateRepositoryContext creates a Level 2 repository context
func (g *ContextGenerator) CreateRepositoryContext(repoName string) Context {
	id := makeContextID(ResourceTypeRepository, repoName)
	parentID := makeContextID(ResourceTypeSource, "")
	return Context{
		Id:           id,
		Name:         fmt.Sprintf("repository:%s", repoName),
		Level:        2,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeRepository,
		Title:        ptrString(repoName),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"repo": repoName,
			},
		},
	}
}

// CreatePRContext creates a Level 3 pull request context
func (g *ContextGenerator) CreatePRContext(repoName string, prNumber int) Context {
	id := makeContextID(ResourceTypePullRequest, fmt.Sprintf("%d", prNumber))
	parentID := makeContextID(ResourceTypeRepository, repoName)
	return Context{
		Id:           id,
		Name:         fmt.Sprintf("PR #%d", prNumber),
		Level:        3,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypePullRequest,
		Title:        ptrString(fmt.Sprintf("PR #%d", prNumber)),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"repo":      repoName,
				"pr_number": fmt.Sprintf("%d", prNumber),
			},
		},
	}
}

// CreateIssueContext creates a Level 3 issue context
func (g *ContextGenerator) CreateIssueContext(repoName string, issueNumber int) Context {
	id := makeContextID(ResourceTypeIssue, fmt.Sprintf("%d", issueNumber))
	parentID := makeContextID(ResourceTypeRepository, repoName)
	return Context{
		Id:           id,
		Name:         fmt.Sprintf("Issue #%d", issueNumber),
		Level:        3,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeIssue,
		Title:        ptrString(fmt.Sprintf("Issue #%d", issueNumber)),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"repo":         repoName,
				"issue_number": fmt.Sprintf("%d", issueNumber),
			},
		},
	}
}
