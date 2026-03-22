//go:build wasip1

package main

import (
	"encoding/json"
	"fmt"
	"google-calendar-connector/internal/auth"
	"google-calendar-connector/internal/core"
	"google-calendar-connector/internal/enrich"
	"net/url"

	"github.com/extism/go-pdk"
)

// EnrichContext enriches the given context with Google Calendar API data
func EnrichContext(input EnrichRequest) (EnrichResponse, error) {
	logger.Info(fmt.Sprintf("EnrichContext: enriching context %s", input.Context.Id))

	config, ok := input.Config.(map[string]any)
	if !ok {
		return EnrichResponse{}, fmt.Errorf("invalid configuration format")
	}

	client, err := auth.NewClient(config)
	if err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to create auth client: %w", err)
	}

	contextType := input.Context.ResourceType

	enrichmentParams, err := extractEnrichmentParams(input.Context.Metadata)
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("No enrichment params for context %s, skipping", input.Context.Id))
		return EnrichResponse{Context: input.Context}, nil
	}

	httpClient := &enrichHTTPClient{client: client}

	enricher, err := enrich.NewContextEnricher(httpClient, contextType, enrichmentParams, logger)
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

// enrichHTTPClient implements enrich.HTTPClient using the auth.Client
type enrichHTTPClient struct {
	client auth.Client
}

func (c *enrichHTTPClient) FetchCalendarDetail(calendarID string) (*enrich.CalendarDetailResponse, error) {
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/calendar/v3/users/me/calendarList/%s",
		url.PathEscape(calendarID),
	)

	body, status, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("calendar detail request failed: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("calendar detail API error (status %d): %s", status, string(body))
	}
	pdk.Log(pdk.LogDebug, fmt.Sprintf("FetchCalendarDetail[%s]: status=%d", calendarID, status))

	var resp enrich.CalendarDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse calendar detail response: %w", err)
	}
	return &resp, nil
}

func (c *enrichHTTPClient) FetchEventDetail(calendarID, eventID string) (*core.Event, error) {
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/calendar/v3/calendars/%s/events/%s",
		url.PathEscape(calendarID),
		url.PathEscape(eventID),
	)

	body, status, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("event detail request failed: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("event detail API error (status %d): %s", status, string(body))
	}
	pdk.Log(pdk.LogDebug, fmt.Sprintf("FetchEventDetail[%s/%s]: status=%d", calendarID, eventID, status))

	var resp core.Event
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse event detail response: %w", err)
	}
	return &resp, nil
}
