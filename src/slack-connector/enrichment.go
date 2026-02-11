package main

import (
	"encoding/json"
	"fmt"
	"slack-connector/internal/core"
	"slack-connector/internal/enrich"

	"github.com/extism/go-pdk"
)

// EnrichContext enriches the given context with Slack API data
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

func (c *enrichHTTPClient) FetchChannel(token, channelID string) (map[string]any, error) {
	// Call Slack API: GET /conversations.info?channel={channel_id}
	url := fmt.Sprintf("%s/conversations.info?channel=%s", core.SlackAPIBaseURL, channelID)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "Bearer "+token)
	req.SetHeader("Content-Type", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("Slack API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check if API call was successful
	ok, _ := apiResp["ok"].(bool)
	if !ok {
		errorMsg := getStringValue(apiResp, "error")
		return nil, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return apiResp, nil
}

func (c *enrichHTTPClient) FetchThread(token, channelID, threadTS string) (map[string]any, error) {
	// Call Slack API: GET /conversations.replies?channel={channel_id}&ts={thread_ts}&limit=1
	url := fmt.Sprintf("%s/conversations.replies?channel=%s&ts=%s&limit=1", core.SlackAPIBaseURL, channelID, threadTS)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "Bearer "+token)
	req.SetHeader("Content-Type", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("Slack API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	ok, _ := apiResp["ok"].(bool)
	if !ok {
		errorMsg := getStringValue(apiResp, "error")
		return nil, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return apiResp, nil
}
