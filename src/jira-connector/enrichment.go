package main

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/core"
	"jira-connector/internal/enrich"

	"github.com/extism/go-pdk"
)

// EnrichContext enriches the given context with Jira API data
func EnrichContext(input EnrichRequest) (EnrichResponse, error) {
	logger.Info(fmt.Sprintf("EnrichContext: Enriching context %s", input.Context.Id))

	contextType := input.Context.ResourceType

	enrichmentParams, err := extractEnrichmentParams(input.Context.Metadata)
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("No enrichment params for context %s, skipping", input.Context.Id))
		return EnrichResponse{
			Context: input.Context,
		}, nil
	}

	enricher, err := enrich.NewContextEnricher(&enrichHTTPClient{}, contextType, input.Config, enrichmentParams, logger)
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
func extractEnrichmentParams(metadata any) (any, error) {
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

	return enrichmentParams, nil
}

type enrichHTTPClient struct{}

func (c *enrichHTTPClient) FetchProject(cloudID, email, apiToken, projectID string) (*enrich.JiraProjectResponse, error) {
	url := fmt.Sprintf("%s/%s/rest/api/3/project/%s", core.JiraAPIBase, cloudID, projectID)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", basicAuthHeader(email, apiToken))
	req.SetHeader("Accept", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("Jira API error (status %d): %s", res.Status(), body)
	}

	var apiResp enrich.JiraProjectResponse
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &apiResp, nil
}

func (c *enrichHTTPClient) FetchIssue(cloudID, email, apiToken, issueID string) (*enrich.JiraIssueResponse, error) {
	url := fmt.Sprintf("%s/%s/rest/api/3/issue/%s", core.JiraAPIBase, cloudID, issueID)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", basicAuthHeader(email, apiToken))
	req.SetHeader("Accept", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("Jira API error (status %d): %s", res.Status(), body)
	}

	var apiResp enrich.JiraIssueResponse
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &apiResp, nil
}
