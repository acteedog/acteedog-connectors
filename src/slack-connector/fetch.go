package main

import (
	"encoding/json"
	"fmt"
	"net/url"
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

	botToken, ok := config["bot_token"].(string)
	if !ok || botToken == "" {
		return FetchResponse{}, fmt.Errorf("missing bot_token")
	}

	userID, ok := config["user_id"].(string)
	if !ok || userID == "" {
		return FetchResponse{}, fmt.Errorf("missing user_id")
	}

	workspaceURL, ok := config["workspace_url"].(string)
	if !ok || workspaceURL == "" {
		return FetchResponse{}, fmt.Errorf("missing workspace_url")
	}

	targetDate := input.Params.TargetDate
	pdk.Log(pdk.LogInfo, fmt.Sprintf("Fetching messages for user: %s, date: %s", userID, targetDate))

	queryDate, err := formatDateForQuery(targetDate)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("invalid target date: %w", err)
	}

	allMessages, err := fetchAllMessages(userID, botToken, queryDate)
	if err != nil {
		return FetchResponse{}, err
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Fetched %d messages", len(allMessages)))

	gen := NewContextGenerator(workspaceURL)
	activities := []Activity{}
	for _, message := range allMessages {
		activity, err := transformMessage(message, gen)
		if err != nil {
			pdk.Log(pdk.LogWarn, fmt.Sprintf("Skipping message: %s", err.Error()))
			continue
		}
		if activity != nil {
			activities = append(activities, *activity)
		}
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Transformed %d activities", len(activities)))

	return FetchResponse{
		Activities: activities,
	}, nil
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

// fetchAllMessages fetches messages from Slack API with pagination
func fetchAllMessages(userID, token, queryDate string) ([]map[string]any, error) {
	allMessages := []map[string]any{}

	// Slack search.messages API pagination (max 100 pages as per YAML)
	for page := 1; page <= 100; page++ {
		query := fmt.Sprintf("from:@%s on:%s", userID, queryDate)
		url := fmt.Sprintf("%s/search.messages?query=%s&count=100&page=%d", SlackAPIBaseURL, url.QueryEscape(query), page)

		pdk.Log(pdk.LogDebug, fmt.Sprintf("Fetching page %d: %s", page, url))

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

		ok, _ := apiResp["ok"].(bool)
		if !ok {
			errorMsg := getStringValue(apiResp, "error")
			return nil, fmt.Errorf("Slack API error: %s", errorMsg)
		}

		messagesObj, ok := apiResp["messages"].(map[string]any)
		if !ok {
			pdk.Log(pdk.LogDebug, "No messages object in response, stopping pagination")
			break
		}

		matchesInterface, ok := messagesObj["matches"]
		if !ok {
			pdk.Log(pdk.LogDebug, "No matches in messages object, stopping pagination")
			break
		}

		matches, ok := matchesInterface.([]any)
		if !ok {
			pdk.Log(pdk.LogDebug, "Matches is not an array, stopping pagination")
			break
		}

		if len(matches) == 0 {
			pdk.Log(pdk.LogDebug, "Empty matches array, stopping pagination")
			break
		}

		for _, match := range matches {
			if msg, ok := match.(map[string]any); ok {
				allMessages = append(allMessages, msg)
			}
		}

		pdk.Log(pdk.LogDebug, fmt.Sprintf("Page %d: %d messages fetched", page, len(matches)))

		paging, ok := messagesObj["paging"].(map[string]any)
		if !ok {
			break
		}

		pages, _ := paging["pages"].(float64)
		currentPage, _ := paging["page"].(float64)

		if currentPage >= pages {
			pdk.Log(pdk.LogDebug, "Reached last page, stopping pagination")
			break
		}
	}

	return allMessages, nil
}

// transformMessage transforms a Slack message to an Activity
func transformMessage(message map[string]any, gen *ContextGenerator) (*Activity, error) {
	ts := getStringValue(message, "ts")
	if ts == "" {
		return nil, fmt.Errorf("message missing ts field")
	}

	text := getStringValue(message, "text")
	permalink := getStringValue(message, "permalink")
	username := getStringValue(message, "username")
	team := getStringValue(message, "team")

	channelObj, ok := message["channel"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("message missing channel object")
	}

	channelID := getStringValue(channelObj, "id")
	channelName := getStringValue(channelObj, "name")

	if channelID == "" || channelName == "" {
		return nil, fmt.Errorf("channel missing id or name")
	}

	// Determine thread_ts (from permalink or fallback to ts)
	threadTS := parseThreadTS(permalink)
	if threadTS == "" {
		threadTS = ts // Standalone message - use its own timestamp
	}

	// Convert timestamp to ISO 8601
	timestamp, err := convertSlackTSToTime(ts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert timestamp: %w", err)
	}

	sourceContext := gen.CreateSourceContext()
	channelContext := gen.CreateChannelContext(channelID, channelName)
	threadContext := gen.CreateThreadContext(channelID, threadTS)

	title := fmt.Sprintf("Message in #%s", channelName)
	description := text

	activity := Activity{
		Id:           makeActivityID(ts),
		Timestamp:    timestamp.Format(time.RFC3339),
		Source:       "Slack",
		ActivityType: "message",
		Title:        title,
		Description:  &description,
		Url:          &permalink,
		Metadata: map[string]any{
			"channel_id":   channelID,
			"channel_name": channelName,
			"user":         username,
			"thread_ts":    threadTS,
			"team":         team,
		},
		Contexts: []Context{
			sourceContext,
			channelContext,
			threadContext,
		},
	}

	return &activity, nil
}

// ============================================================================
// Test Export Functions
// ============================================================================

//go:export TransformMessage
func TransformMessage() int32 {
	input := pdk.Input()

	var message map[string]any
	if err := json.Unmarshal(input, &message); err != nil {
		pdk.SetError(err)
		return 1
	}

	gen := NewContextGenerator("test-workspace.slack.com")

	activity, err := transformMessage(message, gen)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	output, err := json.Marshal(activity)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}
