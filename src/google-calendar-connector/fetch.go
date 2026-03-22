//go:build wasip1

package main

import (
	"encoding/json"
	"fmt"
	"google-calendar-connector/internal/auth"
	"google-calendar-connector/internal/core"
	"google-calendar-connector/internal/fetch"
	"net/url"

	"github.com/extism/go-pdk"
)

// FetchActivities fetches Google Calendar events as activities
func FetchActivities(input FetchRequest) (FetchResponse, error) {
	logger.Info(fmt.Sprintf("FetchActivities: fetching for date %s", input.Params.TargetDate))

	config, ok := input.Config.(map[string]any)
	if !ok {
		return FetchResponse{}, fmt.Errorf("invalid configuration format")
	}

	client, err := auth.NewClient(config)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("failed to create auth client: %w", err)
	}

	targetEmail, _ := config["target_email"].(string)
	httpClient := &fetchHTTPClient{client: client}

	fetcher, err := fetch.NewActivityFetcher(httpClient, input.Params.TargetDate, targetEmail, logger)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("failed to create activity fetcher: %w", err)
	}

	activities, err := fetcher.FetchActivities()
	if err != nil {
		return FetchResponse{}, fmt.Errorf("failed to fetch activities: %w", err)
	}

	result := make([]Activity, 0, len(activities))
	for _, a := range activities {
		result = append(result, convertActivity(a))
	}

	return FetchResponse{Activities: result}, nil
}

func convertActivity(a *fetch.Activity) Activity {
	return Activity{
		Id:           a.Id,
		Timestamp:    a.Timestamp,
		Source:       a.Source,
		ActivityType: a.ActivityType,
		Title:        a.Title,
		Description:  a.Description,
		Url:          a.Url,
		Metadata:     a.Metadata,
		Contexts:     convertContexts(a.Contexts),
	}
}

func convertContexts(contexts []*core.Context) []Context {
	result := make([]Context, 0, len(contexts))
	for _, c := range contexts {
		result = append(result, convertContext(c))
	}
	return result
}

func convertContext(c *core.Context) Context {
	return Context{
		ConnectorId:  c.ConnectorId,
		CreatedAt:    c.CreatedAt,
		Description:  c.Description,
		Id:           c.Id,
		Metadata:     c.Metadata,
		Name:         c.Name,
		ParentId:     c.ParentId,
		ResourceType: c.ResourceType,
		Title:        c.Title,
		UpdatedAt:    c.UpdatedAt,
		Url:          c.Url,
	}
}

// fetchHTTPClient implements fetch.HTTPClient using the auth.Client
type fetchHTTPClient struct {
	client auth.Client
}

func (c *fetchHTTPClient) FetchCalendarList() (*fetch.CalendarListResponse, error) {
	apiURL := "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	body, status, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("calendar list request failed: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("calendar list API error (status %d): %s", status, string(body))
	}
	pdk.Log(pdk.LogDebug, fmt.Sprintf("FetchCalendarList: status=%d", status))

	var resp fetch.CalendarListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse calendar list response: %w", err)
	}
	return &resp, nil
}

func (c *fetchHTTPClient) FetchEvents(calendarID, timeMin, timeMax string) (*fetch.EventListResponse, error) {
	params := url.Values{}
	params.Set("timeMin", timeMin)
	params.Set("timeMax", timeMax)
	params.Set("singleEvents", "true")
	params.Set("orderBy", "startTime")

	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/calendar/v3/calendars/%s/events?%s",
		url.PathEscape(calendarID),
		params.Encode(),
	)

	body, status, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("events request failed: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("events API error (status %d): %s", status, string(body))
	}
	pdk.Log(pdk.LogDebug, fmt.Sprintf("FetchEvents[%s]: status=%d", calendarID, status))

	var resp fetch.EventListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse events response: %w", err)
	}
	return &resp, nil
}

