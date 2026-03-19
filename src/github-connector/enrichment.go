package main

import (
	"encoding/json"
	"fmt"
	"github-connector/internal/auth"
	"github-connector/internal/core"
	"github-connector/internal/enrich"
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
		}, nil
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return EnrichResponse{}, fmt.Errorf("invalid configuration format")
	}

	authClient, err := auth.NewClient(config)
	if err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	enricher, err := enrich.NewContextEnricher(&enrichHTTPClient{authClient: authClient}, contextType, config, enrichmentParams, logger)
	if err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to create context enricher: %w", err)
	}

	enrichedContext, err := enricher.EnrichContext(fromContext(input.Context))
	if err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to enrich context: %w", err)
	}

	return EnrichResponse{
		Context: convertContext(enrichedContext),
	}, nil
}

func fromContext(context Context) *core.Context {
	return &core.Context{
		ConnectorId:  context.ConnectorId,
		CreatedAt:    context.CreatedAt,
		Description:  context.Description,
		Id:           context.Id,
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

type enrichHTTPClient struct {
	authClient auth.Client
}

func (c *enrichHTTPClient) FetchRepository(repo string) (map[string]any, error) {
	url := fmt.Sprintf("%s/repos/%s", core.GithubAPIBaseURL, repo)
	body, status, err := c.authClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("GitHub API error (status %d): %s", status, string(body))
	}

	var apiResp map[string]any
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}

func (c *enrichHTTPClient) FetchPullRequest(repo, number string) (map[string]any, error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%s", core.GithubAPIBaseURL, repo, number)
	body, status, err := c.authClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("GitHub API error (status %d): %s", status, string(body))
	}

	var apiResp map[string]any
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}

func (c *enrichHTTPClient) FetchIssue(repo, number string) (map[string]any, error) {
	url := fmt.Sprintf("%s/repos/%s/issues/%s", core.GithubAPIBaseURL, repo, number)
	body, status, err := c.authClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("GitHub API error (status %d): %s", status, string(body))
	}

	var apiResp map[string]any
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}
