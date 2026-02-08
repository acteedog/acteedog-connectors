package main

import (
	"encoding/json"
	"fmt"
	"github-connector/internal/core"
	"github-connector/internal/enrich"

	"github.com/extism/go-pdk"
)

// EnrichContext enriches the given context with data from GitHub API
func EnrichContext(input EnrichRequest) (EnrichResponse, error) {
	logger.Info(fmt.Sprintf("EnrichContext: Enriching context %s", input.Context.Id))

	contextType := input.Context.ResourceType

	enrichmentParams, err := extractEnrichmentParams(input.Context.Metadata)
	if err != nil {
		logger.Warn(fmt.Sprintf("EnrichContext: No enrichment params for context %s, skipping", input.Context.Id))
		return EnrichResponse{
			Context: input.Context,
			Status:  EnrichResponseStatusEnumSuccess,
		}, nil
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return EnrichResponse{}, fmt.Errorf("invalid configuration format")
	}

	enricher, err := enrich.NewContextEnricher(&enrichHTTPClient{}, contextType, config, enrichmentParams, logger)
	if err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to create context enricher: %w", err)
	}

	enrichedContext, err := enricher.EnrichContext(fromContext(input.Context))
	if err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to enrich context: %w", err)
	}

	return EnrichResponse{
		Context: convertContext(enrichedContext),
		Status:  EnrichResponseStatusEnumSuccess,
	}, nil
}

func fromContext(context Context) *core.Context {
	return &core.Context{
		ConnectorId:  context.ConnectorId,
		CreatedAt:    context.CreatedAt,
		Description:  context.Description,
		Id:           context.Id,
		Level:        context.Level,
		Metadata:     context.Metadata,
		Name:         context.Name,
		ParentId:     context.ParentId,
		ResourceType: context.ResourceType,
		Title:        context.Title,
		UpdatedAt:    context.UpdatedAt,
		Url:          context.Url,
	}
}

func convertContext(context *core.Context) Context {
	return Context{
		ConnectorId:  context.ConnectorId,
		CreatedAt:    context.CreatedAt,
		Description:  context.Description,
		Id:           context.Id,
		Level:        context.Level,
		Metadata:     context.Metadata,
		Name:         context.Name,
		ParentId:     context.ParentId,
		ResourceType: context.ResourceType,
		Title:        context.Title,
		UpdatedAt:    context.UpdatedAt,
		Url:          context.Url,
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

type enrichHTTPClient struct{}

func (c *enrichHTTPClient) FetchRepository(token, repo string) (map[string]any, error) {
	url := fmt.Sprintf("%s/repos/%s", GithubAPIBaseURL, repo)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "token "+token)
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/"+ConnectorID)

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("GitHub API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}

func (c *enrichHTTPClient) FetchPullRequest(token, repo, number string) (map[string]any, error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%s", GithubAPIBaseURL, repo, number)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "token "+token)
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/"+ConnectorID)

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("GitHub API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}

func (c *enrichHTTPClient) FetchIssue(token, repo, number string) (map[string]any, error) {
	url := fmt.Sprintf("%s/repos/%s/issues/%s", GithubAPIBaseURL, repo, number)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "token "+token)
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/"+ConnectorID)

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("GitHub API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}
