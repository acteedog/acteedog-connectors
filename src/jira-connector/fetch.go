package main

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/core"
	"jira-connector/internal/fetch"
	"net/url"
	"strings"

	"github.com/extism/go-pdk"
)

// FetchActivities fetches Jira activities for a user on a specific date
func FetchActivities(input FetchRequest) (FetchResponse, error) {
	logger.Info("FetchActivities: Starting Jira activities fetch")

	fetcher, err := fetch.NewActivityFetcher(&fetchHTTPClient{}, input.Config, input.Params.TargetDate, logger)
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

func (c *fetchHTTPClient) FetchIssues(cloudID, email, apiToken string, projectIDs []string, dateFrom, dateTo string) (*fetch.JiraSearchResponse, error) {
	// Build project list for JQL: "10000","10001"
	quotedIDs := make([]string, len(projectIDs))
	for i, id := range projectIDs {
		quotedIDs[i] = `"` + id + `"`
	}
	projectList := strings.Join(quotedIDs, ",")

	jql := fmt.Sprintf(
		`updated >= "%s" AND updated < "%s" AND project IN (%s) ORDER BY created DESC`,
		dateFrom, dateTo, projectList,
	)

	apiURL := fmt.Sprintf(
		"%s/%s/rest/api/3/search/jql?jql=%s&fields=*all&expand=changelog&maxResults=5000",
		core.JiraAPIBase,
		cloudID,
		url.QueryEscape(jql),
	)

	logger.Debug(fmt.Sprintf("Fetching issues: %s", apiURL))

	req := pdk.NewHTTPRequest(pdk.MethodGet, apiURL)
	req.SetHeader("Authorization", basicAuthHeader(email, apiToken))
	req.SetHeader("Accept", "application/json")

	res := req.Send()

	if res.Status() != 200 {
		body := string(res.Body())
		return nil, fmt.Errorf("Jira API error: HTTP %d, body: %s", res.Status(), body)
	}

	var apiResp fetch.JiraSearchResponse
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &apiResp, nil
}
