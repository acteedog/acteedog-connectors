package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slack-connector/internal/core"
	"slack-connector/internal/fetch"
	"time"

	"github.com/extism/go-pdk"
)

// FetchActivities fetches Slack messages for a user on a specific date
func FetchActivities(input FetchRequest) (FetchResponse, error) {
	pdk.Log(pdk.LogInfo, "FetchActivities: Starting Slack messages fetch")

	// Parse configuration
	config, ok := input.Config.(map[string]any)
	if !ok {
		return FetchResponse{}, fmt.Errorf("invalid configuration format")
	}

	fetcher, err := fetch.NewActivityFetcher(&fetchHTTPClient{}, config, input.Params.TargetDate, logger)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("failed to create activity fetcher: %w", err)
	}

	activities, err := fetcher.FetchActivities()
	if err != nil {
		return FetchResponse{}, fmt.Errorf("failed to fetch activities: %w", err)
	}

	return FetchResponse{
		Activities: convertActivities(activities),
	}, nil
}

func convertContexts(contexts []*core.Context) []Context {
	converted := make([]Context, len(contexts))
	for i, ctx := range contexts {
		converted[i] = convertContext(ctx)
	}
	return converted
}

func convertActivities(activities []*fetch.Activity) []Activity {
	converted := make([]Activity, len(activities))
	for i, activity := range activities {
		converted[i] = convertActivity(activity)
	}
	return converted
}

func convertActivity(activity *fetch.Activity) Activity {
	return Activity{
		ActivityType: activity.ActivityType,
		Contexts:     convertContexts(activity.Contexts),
		Description:  activity.Description,
		Id:           activity.Id,
		Metadata:     activity.Metadata,
		Source:       activity.Source,
		Timestamp:    activity.Timestamp,
		Title:        activity.Title,
		Url:          activity.Url,
	}
}

type fetchHTTPClient struct{}

func (c *fetchHTTPClient) FetchMessages(token, userID, targetDate string, page int) (map[string]any, error) {
	queryDate, err := formatDateForQuery(targetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target date: %w", err)
	}

	query := fmt.Sprintf("from:@%s on:%s", userID, queryDate)
	url := fmt.Sprintf("%s/search.messages?query=%s&count=100&page=%d", core.SlackAPIBaseURL, url.QueryEscape(query), page)

	logger.Debug(fmt.Sprintf("Fetching page %d: %s", page, url))

	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "Bearer "+token)
	req.SetHeader("Content-Type", "application/json")

	res := req.Send()

	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("Slack API error: HTTP %d, body: %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return apiResp, nil
}

// formatDateForQuery formats the target date for Slack search query
// Input: "2025-12-13T00:00:00Z" or "2025-12-13"
// Output: "2025-12-13"
func formatDateForQuery(targetDate string) (string, error) {
	var t time.Time
	var err error

	// Try RFC3339 format first
	t, err = time.Parse(time.RFC3339, targetDate)
	if err != nil {
		// Fallback to date-only format
		t, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return "", err
		}
	}

	// Return in YYYY-MM-DD format
	return t.Format("2006-01-02"), nil
}
