package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/extism/go-pdk"
)

// EnrichContext enriches the given context with data from GitHub API
func EnrichContext(input EnrichRequest) (EnrichResponse, error) {
	pdk.Log(pdk.LogInfo, fmt.Sprintf("EnrichContext: Enriching context %s", input.Context.Id))

	contextType := input.Context.ResourceType

	enrichmentParams, err := extractEnrichmentParams(input.Context.Metadata)
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("No enrichment params for context %s, skipping", input.Context.Id))
		return EnrichResponse{
			Context: input.Context,
			Status:  EnrichResponseStatusEnumSuccess,
		}, nil
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return EnrichResponse{}, fmt.Errorf("invalid configuration format")
	}

	token, ok := config["credential_personal_access_token"].(string)
	if !ok || token == "" {
		return EnrichResponse{}, fmt.Errorf("missing personal access token")
	}

	switch contextType {
	case ResourceTypeSource:
		context := input.Context
		context.Title = ptrString("GitHub")
		context.Description = ptrString("Github is a code hosting platform for version control and collaboration.")
		context.Url = ptrString("https://github.com")

		return EnrichResponse{
			Context: context,
			Status:  EnrichResponseStatusEnumSuccess,
		}, nil
	case ResourceTypeRepository:
		return enrichRepository(input.Context, enrichmentParams, token)
	case ResourceTypePullRequest:
		return enrichPullRequest(input.Context, enrichmentParams, token)
	case ResourceTypeIssue:
		return enrichIssue(input.Context, enrichmentParams, token)
	default:
		return EnrichResponse{}, fmt.Errorf("github-connector: unknown context type: %s (full id: %s)", contextType, input.Context.Id)
	}
}

// extractEnrichmentParams extracts enrichment_params from context metadata
func extractEnrichmentParams(metadata any) (map[string]any, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata is nil")
	}

	metadataMap, ok := metadata.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("metadata is not a map")
	}

	enrichmentParams, ok := metadataMap["enrichment_params"]
	if !ok {
		return nil, fmt.Errorf("enrichment_params not found")
	}

	params, ok := enrichmentParams.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("enrichment_params is not a map")
	}

	return params, nil
}

// enrichRepository enriches repository context with GitHub API data
func enrichRepository(context Context, params map[string]any, token string) (EnrichResponse, error) {
	repo, ok := params["repo"].(string)
	if !ok || repo == "" {
		return EnrichResponse{}, fmt.Errorf("repo not found in enrichment_params")
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Enriching repository: %s", repo))

	// Call GitHub API: GET /repos/{owner}/{repo}
	url := fmt.Sprintf("%s/repos/%s", GithubAPIBaseURL, repo)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "token "+token)
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/"+ConnectorVersion)

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return EnrichResponse{}, fmt.Errorf("GitHub API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to parse API response: %w", err)
	}

	return applyRepositoryEnrichment(context, apiResp), nil
}

// applyRepositoryEnrichment applies repository API response to context
func applyRepositoryEnrichment(context Context, apiResp map[string]any) EnrichResponse {
	title := fmt.Sprintf("Repository: %s", getStringValue(apiResp, "full_name"))
	description := getStringValue(apiResp, "description")
	url := getStringValue(apiResp, "html_url")
	createdAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "created_at"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse created_at: %v", err))
	} else {
		createdAt = createdAt.UTC()
		context.CreatedAt = &createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "updated_at"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse updated_at: %v", err))
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

	return EnrichResponse{
		Context: context,
		Status:  EnrichResponseStatusEnumSuccess,
	}
}

// enrichPullRequest enriches pull request context with GitHub API data
func enrichPullRequest(context Context, params map[string]any, token string) (EnrichResponse, error) {
	repo, ok := params["repo"].(string)
	if !ok || repo == "" {
		return EnrichResponse{}, fmt.Errorf("repo not found in enrichment_params")
	}

	// Support both string (new) and float64 (old) for backward compatibility
	var prNumber float64
	if prNumberStr, ok := params["pr_number"].(string); ok {
		var err error
		prNumber, err = strconv.ParseFloat(prNumberStr, 64)
		if err != nil {
			return EnrichResponse{}, fmt.Errorf("invalid pr_number format: %v", err)
		}
	} else if prNumberFloat, ok := params["pr_number"].(float64); ok {
		prNumber = prNumberFloat
	} else {
		return EnrichResponse{}, fmt.Errorf("pr_number not found in enrichment_params")
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Enriching PR: %s #%.0f", repo, prNumber))

	// Call GitHub API: GET /repos/{owner}/{repo}/pulls/{pr_number}
	url := fmt.Sprintf("%s/repos/%s/pulls/%.0f", GithubAPIBaseURL, repo, prNumber)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "token "+token)
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/"+ConnectorVersion)

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return EnrichResponse{}, fmt.Errorf("GitHub API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to parse API response: %w", err)
	}

	return applyPullRequestEnrichment(context, apiResp), nil
}

// applyPullRequestEnrichment applies pull request API response to context
func applyPullRequestEnrichment(context Context, apiResp map[string]any) EnrichResponse {
	prTitle := getStringValue(apiResp, "title")
	prDescription := getStringValue(apiResp, "body")
	prUrl := getStringValue(apiResp, "html_url")
	createdAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "created_at"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse created_at: %v", err))
	} else {
		createdAt = createdAt.UTC()
		context.CreatedAt = &createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "updated_at"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse updated_at: %v", err))
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

	return EnrichResponse{
		Context: context,
		Status:  EnrichResponseStatusEnumSuccess,
	}
}

// enrichIssue enriches issue context with GitHub API data
func enrichIssue(context Context, params map[string]any, token string) (EnrichResponse, error) {
	repo, ok := params["repo"].(string)
	if !ok || repo == "" {
		return EnrichResponse{}, fmt.Errorf("repo not found in enrichment_params")
	}

	// Support both string (new) and float64 (old) for backward compatibility
	var issueNumber float64
	if issueNumberStr, ok := params["issue_number"].(string); ok {
		var err error
		issueNumber, err = strconv.ParseFloat(issueNumberStr, 64)
		if err != nil {
			return EnrichResponse{}, fmt.Errorf("invalid issue_number format: %v", err)
		}
	} else if issueNumberFloat, ok := params["issue_number"].(float64); ok {
		issueNumber = issueNumberFloat
	} else {
		return EnrichResponse{}, fmt.Errorf("issue_number not found in enrichment_params")
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Enriching Issue: %s #%.0f", repo, issueNumber))

	// Call GitHub API: GET /repos/{owner}/{repo}/issues/{issue_number}
	url := fmt.Sprintf("%s/repos/%s/issues/%.0f", GithubAPIBaseURL, repo, issueNumber)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "token "+token)
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/"+ConnectorVersion)

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return EnrichResponse{}, fmt.Errorf("GitHub API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to parse API response: %w", err)
	}

	return applyIssueEnrichment(context, apiResp), nil
}

// applyIssueEnrichment applies issue API response to context
func applyIssueEnrichment(context Context, apiResp map[string]any) EnrichResponse {
	// Update context with enriched data
	issueTitle := getStringValue(apiResp, "title")
	issueDescription := getStringValue(apiResp, "body")
	issueUrl := getStringValue(apiResp, "html_url")
	createdAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "created_at"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse created_at: %v", err))
	} else {
		createdAt = createdAt.UTC()
		context.CreatedAt = &createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339, getStringValue(apiResp, "updated_at"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse updated_at: %v", err))
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

	return EnrichResponse{
		Context: context,
		Status:  EnrichResponseStatusEnumSuccess,
	}
}

// ============================================================================
// Test Exports
// ============================================================================

//go:export ApplyRepositoryEnrichment
func ApplyRepositoryEnrichment() int32 {
	input := pdk.Input()

	var apiResp map[string]any
	if err := json.Unmarshal(input, &apiResp); err != nil {
		pdk.SetError(err)
		return 1
	}

	gen := NewContextGenerator()
	context := gen.CreateRepositoryContext("testorg/testrepo")

	response := applyRepositoryEnrichment(context, apiResp)

	output, err := json.Marshal(response.Context)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}

//go:export ApplyPullRequestEnrichment
func ApplyPullRequestEnrichment() int32 {
	input := pdk.Input()

	var apiResp map[string]any
	if err := json.Unmarshal(input, &apiResp); err != nil {
		pdk.SetError(err)
		return 1
	}

	gen := NewContextGenerator()
	context := gen.CreatePRContext("testorg/testrepo", 52742)

	response := applyPullRequestEnrichment(context, apiResp)

	output, err := json.Marshal(response.Context)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}

//go:export ApplyIssueEnrichment
func ApplyIssueEnrichment() int32 {
	input := pdk.Input()

	var apiResp map[string]any
	if err := json.Unmarshal(input, &apiResp); err != nil {
		pdk.SetError(err)
		return 1
	}

	gen := NewContextGenerator()
	context := gen.CreateIssueContext("ymtdzzz/otel-tui", 340)

	response := applyIssueEnrichment(context, apiResp)

	output, err := json.Marshal(response.Context)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}
