package main

import (
	"encoding/json"
	"fmt"
	"github-connector/internal/auth"
	"github-connector/internal/core"
	"github-connector/internal/fetch"

	"github.com/extism/go-pdk"
)

// FetchActivities fetches GitHub activities based on the input configuration and parameters
func FetchActivities(input FetchRequest) (FetchResponse, error) {
	logger.Info("FetchActivities: Starting GitHub events fetch")

	config, ok := input.Config.(map[string]any)
	if !ok {
		return FetchResponse{}, fmt.Errorf("invalid configuration format")
	}

	authClient, err := auth.NewClient(config)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	fetcher, err := fetch.NewActivityFetcher(&fetchHTTPClient{authClient: authClient}, config, input.Params.TargetDate, logger)
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

type fetchHTTPClient struct {
	authClient auth.Client
}

func (c *fetchHTTPClient) FetchActivities(username string, page int) ([]map[string]any, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events?per_page=100&page=%d", username, page)

	pdk.Log(pdk.LogDebug, fmt.Sprintf("Fetching page %d: %s", page, url))

	body, status, err := c.authClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("GitHub API error: HTTP %d", status)
	}

	var events []map[string]any
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	return events, nil
}
