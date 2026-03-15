package enrich

import (
	"fmt"
	"jira-connector/internal/core"
)

// ContextEnricher defines the structure for enriching context data from Jira
type ContextEnricher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewContextEnricher creates a new ContextEnricher instance
func NewContextEnricher(httpClient HTTPClient, contextType string, cfg, params any, logger core.Logger) (*ContextEnricher, error) {
	config, err := newConfig(contextType, cfg, params)
	if err != nil {
		return nil, err
	}

	return &ContextEnricher{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}, nil
}

// EnrichContext enriches the given context with additional data from Jira
func (e *ContextEnricher) EnrichContext(context *core.Context) (*core.Context, error) {
	e.logger.Info("Starting to enrich context")

	switch e.config.contextType {
	case core.ResourceTypeSource:
		title := "Jira"
		description := "Activity source from Jira"
		url := fmt.Sprintf("%s/%s", core.JiraAPIBase, e.config.CloudID)
		context.Title = &title
		context.Description = &description
		context.Url = &url
		return context, nil
	case core.ResourceTypeProject:
		return e.enrichProject(context)
	case core.ResourceTypeIssue:
		return e.enrichIssue(context)
	default:
		return nil, fmt.Errorf("unsupported context type: %s", e.config.contextType)
	}
}

func (e *ContextEnricher) enrichProject(context *core.Context) (*core.Context, error) {
	projectID := e.config.enrichmentParams.ProjectID
	if projectID == "" {
		return nil, fmt.Errorf("project_id not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching project: %s", projectID))

	response, err := e.httpClient.FetchProject(e.config.CloudID, e.config.Email, e.config.APIToken, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project data: %w", err)
	}

	return e.applyProjectEnrichment(context, response)
}

func (e *ContextEnricher) applyProjectEnrichment(context *core.Context, project *JiraProjectResponse) (*core.Context, error) {
	title := project.Name
	description := project.Description
	url := fmt.Sprintf("https://%s.atlassian.net/jira/software/projects/%s/boards", e.config.SiteSubdomain, project.Key)

	context.Title = &title
	context.Description = &description
	context.Url = &url

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["key"] = project.Key
	metadataMap["project_type_key"] = project.ProjectTypeKey

	context.Metadata = metadataMap

	return context, nil
}

func (e *ContextEnricher) enrichIssue(context *core.Context) (*core.Context, error) {
	issueID := e.config.enrichmentParams.IssueID
	if issueID == "" {
		return nil, fmt.Errorf("issue_id not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching issue: %s", issueID))

	response, err := e.httpClient.FetchIssue(e.config.CloudID, e.config.Email, e.config.APIToken, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue data: %w", err)
	}

	return e.applyIssueEnrichment(context, response)
}

func (e *ContextEnricher) applyIssueEnrichment(context *core.Context, issue *JiraIssueResponse) (*core.Context, error) {
	f := issue.Fields

	title := f.Summary
	description := f.Description.PlainText()
	url := fmt.Sprintf("https://%s.atlassian.net/browse/%s", e.config.SiteSubdomain, issue.Key)

	context.Title = &title
	context.Description = &description
	context.Url = &url

	if f.Created != "" {
		if t, err := core.ParseJiraTime(f.Created); err == nil {
			utc := t.UTC()
			context.CreatedAt = &utc
		}
	}
	if f.Updated != "" {
		if t, err := core.ParseJiraTime(f.Updated); err == nil {
			utc := t.UTC()
			context.UpdatedAt = &utc
		}
	}

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	if f.IssueType != nil {
		metadataMap["issue_type"] = f.IssueType.Name
	}
	if f.Status != nil {
		metadataMap["status"] = f.Status.Name
	}
	if f.Priority != nil {
		metadataMap["priority"] = f.Priority.Name
	}
	if f.Creator != nil {
		metadataMap["creator"] = f.Creator.EmailAddress
	}
	metadataMap["created"] = f.Created
	metadataMap["updated"] = f.Updated

	context.Metadata = metadataMap

	return context, nil
}
