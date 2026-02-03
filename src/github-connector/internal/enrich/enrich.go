package enrich

import (
	"fmt"
	"github-connector/internal/core"
	"time"
)

// ContextEnricher defines the structure for enriching context data from GitHub
type ContextEnricher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewContextEnricher creates a new ContextEnricher instance
func NewContextEnricher(httpClient HTTPClient, contextType string, cfg, params map[string]any, logger core.Logger) (*ContextEnricher, error) {
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

// EnrichContext enriches the given context with additional data from GitHub
func (e *ContextEnricher) EnrichContext(context *core.Context) (*core.Context, error) {
	e.logger.Info("Starting to enrich context")

	switch e.config.contextType {
	case core.ResourceTypeSource:
		context.Title = ptrString("GitHub")
		context.Description = ptrString("Github is a code hosting platform for version control and collaboration.")
		context.Url = ptrString("https://github.com")

		return context, nil
	case core.ResourceTypeRepository:
		return e.enrichRepository(context)
	case core.ResourceTypePullRequest:
		return e.enrichPullRequest(context)
	case core.ResourceTypeIssue:
		return e.enrichIssue(context)
	default:
		return nil, fmt.Errorf("unsupported context type: %s", e.config.contextType)
	}
}

func (e *ContextEnricher) enrichRepository(context *core.Context) (*core.Context, error) {
	repo, ok := e.config.enrichmentParams["repo"].(string)
	if !ok || repo == "" {
		return nil, fmt.Errorf("repo not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching repository: %s", repo))

	response, err := e.httpClient.FetchRepository(e.config.token, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository data: %w", err)
	}

	return e.applyRepositoryEnrichment(context, response)
}

func (e *ContextEnricher) applyRepositoryEnrichment(context *core.Context, apiResp map[string]any) (*core.Context, error) {
	title := fmt.Sprintf("Repository: %s", getStringValue(apiResp, "full_name"))
	description := getStringValue(apiResp, "description")
	url := getStringValue(apiResp, "html_url")
	createdAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "created_at"))
	if err != nil {
		return nil, err
	} else {
		createdAt = createdAt.UTC()
		context.CreatedAt = &createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "updated_at"))
	if err != nil {
		return nil, err
	} else {
		updatedAt = updatedAt.UTC()
		context.UpdatedAt = &updatedAt
	}

	context.Title = &title
	context.Description = &description
	context.Url = &url

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["stargazers_count"] = apiResp["stargazers_count"]
	metadataMap["language"] = apiResp["language"]
	metadataMap["topics"] = apiResp["topics"]
	metadataMap["default_branch"] = apiResp["default_branch"]
	metadataMap["visibility"] = apiResp["visibility"]
	metadataMap["forks_count"] = apiResp["forks_count"]
	metadataMap["open_issues_count"] = apiResp["open_issues_count"]
	metadataMap["watchers_count"] = apiResp["watchers_count"]
	metadataMap["homepage"] = apiResp["homepage"]

	context.Metadata = metadataMap

	return context, nil
}

func (e *ContextEnricher) enrichPullRequest(context *core.Context) (*core.Context, error) {
	repo, ok := e.config.enrichmentParams["repo"].(string)
	if !ok || repo == "" {
		return nil, fmt.Errorf("repo not found in enrichment_params")
	}
	number, ok := e.config.enrichmentParams["pr_number"].(string)
	if !ok || number == "" {
		return nil, fmt.Errorf("pr_number not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching pull request: %s #%s", repo, number))

	response, err := e.httpClient.FetchPullRequest(e.config.token, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pull request data: %w", err)
	}

	return e.applyPullRequestEnrichment(context, response)
}

func (e *ContextEnricher) applyPullRequestEnrichment(context *core.Context, apiResp map[string]any) (*core.Context, error) {
	prTitle := getStringValue(apiResp, "title")
	prDescription := getStringValue(apiResp, "body")
	prUrl := getStringValue(apiResp, "html_url")
	createdAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "created_at"))
	if err != nil {
		return nil, err
	} else {
		createdAt = createdAt.UTC()
		context.CreatedAt = &createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "updated_at"))
	if err != nil {
		return nil, err
	} else {
		updatedAt = updatedAt.UTC()
		context.UpdatedAt = &updatedAt
	}

	context.Title = &prTitle
	context.Description = &prDescription
	context.Url = &prUrl

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["state"] = apiResp["state"]
	metadataMap["author"] = getNestedString(apiResp, "user", "login")
	metadataMap["assignees"] = extractLogins(apiResp["assignees"])
	metadataMap["reviewers"] = extractLogins(apiResp["requested_reviewers"])
	metadataMap["labels"] = extractLabelNames(apiResp["labels"])
	metadataMap["base_branch"] = getNestedString(apiResp, "base", "ref")
	metadataMap["head_branch"] = getNestedString(apiResp, "head", "ref")
	metadataMap["milestone"] = getNestedString(apiResp, "milestone", "title")
	metadataMap["additions"] = apiResp["additions"]
	metadataMap["deletions"] = apiResp["deletions"]
	metadataMap["changed_files"] = apiResp["changed_files"]
	metadataMap["commits_count"] = apiResp["commits"]
	metadataMap["merged"] = apiResp["merged"]
	metadataMap["merged_at"] = apiResp["merged_at"]
	metadataMap["merged_by"] = getNestedString(apiResp, "merged_by", "login")

	context.Metadata = metadataMap

	return context, nil
}

func (e *ContextEnricher) enrichIssue(context *core.Context) (*core.Context, error) {
	repo, ok := e.config.enrichmentParams["repo"].(string)
	if !ok || repo == "" {
		return nil, fmt.Errorf("repo not found in enrichment_params")
	}
	number, ok := e.config.enrichmentParams["issue_number"].(string)
	if !ok || number == "" {
		return nil, fmt.Errorf("issue_number not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching issue: %s #%s", repo, number))

	response, err := e.httpClient.FetchIssue(e.config.token, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue data: %w", err)
	}

	return e.applyIssueEnrichment(context, response)
}

func (e *ContextEnricher) applyIssueEnrichment(context *core.Context, apiResp map[string]any) (*core.Context, error) {
	issueTitle := getStringValue(apiResp, "title")
	issueDescription := getStringValue(apiResp, "body")
	issueUrl := getStringValue(apiResp, "html_url")
	createdAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "created_at"))
	if err != nil {
		return nil, err
	} else {
		createdAt = createdAt.UTC()
		context.CreatedAt = &createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "updated_at"))
	if err != nil {
		return nil, err
	} else {
		updatedAt = updatedAt.UTC()
		context.UpdatedAt = &updatedAt
	}

	context.Title = &issueTitle
	context.Description = &issueDescription
	context.Url = &issueUrl

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["state"] = apiResp["state"]
	metadataMap["author"] = getNestedString(apiResp, "user", "login")
	metadataMap["assignees"] = extractLogins(apiResp["assignees"])
	metadataMap["labels"] = extractLabelNames(apiResp["labels"])
	metadataMap["milestone"] = getNestedString(apiResp, "milestone", "title")
	metadataMap["comments"] = apiResp["comments"]

	context.Metadata = metadataMap

	return context, nil
}

// getStringValue safely extracts string value from map
func getStringValue(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getNestedString safely extracts nested string value
func getNestedString(m map[string]any, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - extract string
			return getStringValue(current, key)
		}
		// Navigate deeper
		if nested, ok := current[key].(map[string]any); ok {
			current = nested
		} else {
			return ""
		}
	}
	return ""
}

// extractLogins extracts login names from array of user objects
func extractLogins(usersInterface any) []string {
	if usersInterface == nil {
		return []string{}
	}

	users, ok := usersInterface.([]any)
	if !ok {
		return []string{}
	}

	logins := make([]string, 0, len(users))
	for _, userInterface := range users {
		user, ok := userInterface.(map[string]any)
		if !ok {
			continue
		}
		if login, ok := user["login"].(string); ok {
			logins = append(logins, login)
		}
	}
	return logins
}

// extractLabelNames extracts label names from array of label objects
func extractLabelNames(labelsInterface any) []string {
	if labelsInterface == nil {
		return []string{}
	}

	labels, ok := labelsInterface.([]any)
	if !ok {
		return []string{}
	}

	names := make([]string, 0, len(labels))
	for _, labelInterface := range labels {
		label, ok := labelInterface.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := label["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}
