package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/extism/go-pdk"
)

// EnrichContext enriches the given context with Slack API data
func EnrichContext(input EnrichRequest) (EnrichResponse, error) {
	pdk.Log(pdk.LogInfo, fmt.Sprintf("EnrichContext: Enriching context %s", input.Context.Id))

	contextType := input.Context.ResourceType

	enrichmentParams, err := extractEnrichmentParams(input.Context.Metadata)
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("No enrichment params for context %s, skipping", input.Context.Id))
		return EnrichResponse{
			Context: input.Context,
			Status:  EnrichResponseStatusEnumSuccess,
		}, nil
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return EnrichResponse{}, fmt.Errorf("invalid configuration format")
	}

	botToken, ok := config["bot_token"].(string)
	if !ok || botToken == "" {
		return EnrichResponse{}, fmt.Errorf("missing bot_token")
	}

	workspaceURL, ok := config["workspace_url"].(string)
	if !ok || workspaceURL == "" {
		return EnrichResponse{}, fmt.Errorf("missing workspace_url")
	}

	switch contextType {
	case ResourceTypeSource:
		context := input.Context
		context.Title = ptrString("Slack")
		context.Description = ptrString("Activity source from Slack")
		context.Url = ptrString("https://" + workspaceURL)

		return EnrichResponse{
			Context: context,
			Status:  EnrichResponseStatusEnumSuccess,
		}, nil
	case ResourceTypeChannel:
		return enrichChannel(input.Context, enrichmentParams, botToken, workspaceURL)
	case ResourceTypeThread:
		return enrichThread(input.Context, enrichmentParams, botToken, workspaceURL)
	default:
		return EnrichResponse{}, fmt.Errorf("slack-connector: unknown context type: %s (full id: %s)", contextType, input.Context.Id)
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

// enrichChannel enriches channel context with Slack API data
func enrichChannel(context Context, params map[string]any, token, workspaceURL string) (EnrichResponse, error) {
	channelID, ok := params["channel_id"].(string)
	if !ok || channelID == "" {
		return EnrichResponse{}, fmt.Errorf("channel_id not found in enrichment_params")
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Enriching channel: %s", channelID))

	// Call Slack API: GET /conversations.info?channel={channel_id}
	url := fmt.Sprintf("%s/conversations.info?channel=%s", SlackAPIBaseURL, channelID)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "Bearer "+token)
	req.SetHeader("Content-Type", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return EnrichResponse{}, fmt.Errorf("Slack API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check if API call was successful
	ok, _ = apiResp["ok"].(bool)
	if !ok {
		errorMsg := getStringValue(apiResp, "error")
		return EnrichResponse{}, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return applyChannelEnrichment(context, apiResp, workspaceURL), nil
}

// applyChannelEnrichment applies channel API response to context
func applyChannelEnrichment(context Context, apiResp map[string]any, workspaceURL string) EnrichResponse {
	channelObj, ok := apiResp["channel"].(map[string]any)
	if !ok {
		pdk.Log(pdk.LogWarn, "No channel object in API response")
		return EnrichResponse{
			Context: context,
			Status:  EnrichResponseStatusEnumSuccess,
		}
	}

	name := getStringValue(channelObj, "name")
	topicValue := getNestedString(channelObj, "topic", "value")
	channelID := getStringValue(channelObj, "id")

	title := fmt.Sprintf("#%s", name)
	description := topicValue
	url := fmt.Sprintf("https://%s/archives/%s", workspaceURL, channelID)
	createdAt := time.Unix(getIntValue(channelObj, "created"), 0).UTC()
	updatedAt := time.UnixMilli(getIntValue(channelObj, "updated")).UTC()

	context.Title = &title
	context.Description = &description
	context.Url = &url
	context.CreatedAt = &createdAt
	context.UpdatedAt = &updatedAt

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["name"] = name
	metadataMap["is_private"] = getBoolValue(channelObj, "is_private")
	metadataMap["is_channel"] = getBoolValue(channelObj, "is_channel")
	metadataMap["is_group"] = getBoolValue(channelObj, "is_group")
	metadataMap["is_im"] = getBoolValue(channelObj, "is_im")
	metadataMap["topic"] = topicValue
	metadataMap["purpose"] = getNestedString(channelObj, "purpose", "value")
	metadataMap["context_team_id"] = getStringValue(channelObj, "context_team_id")
	metadataMap["created"] = getIntValue(channelObj, "created")

	context.Metadata = metadataMap

	return EnrichResponse{
		Context: context,
		Status:  EnrichResponseStatusEnumSuccess,
	}
}

// enrichThread enriches thread context with Slack API data
func enrichThread(context Context, params map[string]any, token, workspaceURL string) (EnrichResponse, error) {
	channelID, ok := params["channel_id"].(string)
	if !ok || channelID == "" {
		return EnrichResponse{}, fmt.Errorf("channel_id not found in enrichment_params")
	}

	threadTS, ok := params["thread_ts"].(string)
	if !ok || threadTS == "" {
		return EnrichResponse{}, fmt.Errorf("thread_ts not found in enrichment_params")
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Enriching thread: %s in channel %s", threadTS, channelID))

	// Call Slack API: GET /conversations.replies?channel={channel_id}&ts={thread_ts}&limit=1
	url := fmt.Sprintf("%s/conversations.replies?channel=%s&ts=%s&limit=1", SlackAPIBaseURL, channelID, threadTS)
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", "Bearer "+token)
	req.SetHeader("Content-Type", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		body := string(res.Body())
		return EnrichResponse{}, fmt.Errorf("Slack API error (status %d): %s", res.Status(), body)
	}

	var apiResp map[string]any
	if err := json.Unmarshal(res.Body(), &apiResp); err != nil {
		return EnrichResponse{}, fmt.Errorf("failed to parse API response: %w", err)
	}

	ok, _ = apiResp["ok"].(bool)
	if !ok {
		errorMsg := getStringValue(apiResp, "error")
		return EnrichResponse{}, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return applyThreadEnrichment(context, apiResp, workspaceURL, channelID), nil
}

// applyThreadEnrichment applies thread API response to context
func applyThreadEnrichment(context Context, apiResp map[string]any, workspaceURL, channelID string) EnrichResponse {
	messagesInterface, ok := apiResp["messages"]
	if !ok {
		pdk.Log(pdk.LogWarn, "No messages array in API response")
		return EnrichResponse{
			Context: context,
			Status:  EnrichResponseStatusEnumSuccess,
		}
	}

	messages, ok := messagesInterface.([]any)
	if !ok || len(messages) == 0 {
		pdk.Log(pdk.LogWarn, "Messages array is empty or invalid")
		return EnrichResponse{
			Context: context,
			Status:  EnrichResponseStatusEnumSuccess,
		}
	}

	// Get first message (parent message)
	parentMsg, ok := messages[0].(map[string]any)
	if !ok {
		pdk.Log(pdk.LogWarn, "First message is not a map")
		return EnrichResponse{
			Context: context,
			Status:  EnrichResponseStatusEnumSuccess,
		}
	}

	text := getStringValue(parentMsg, "text")
	parentTS := getStringValue(parentMsg, "ts")

	title := fmt.Sprintf("Thread: %s", text)
	description := text
	// Format: https://{workspace_url}/archives/{channel_id}/p{ts with dots removed}
	url := fmt.Sprintf("https://%s/archives/%s/p%s", workspaceURL, channelID, formatSlackTS(parentTS))
	createdAt, err := parseSlackTS(getStringValue(parentMsg, "thread_ts"))
	if err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse createdAt for thread %s: %v", context.Id, err))
	} else {
		context.CreatedAt = &createdAt
		context.UpdatedAt = &createdAt // Slack does not provide updatedAt for threads
	}

	context.Title = &title
	context.Description = &description
	context.Url = &url

	metadataMap, _ := context.Metadata.(map[string]any)
	if metadataMap == nil {
		metadataMap = make(map[string]any)
	}

	metadataMap["parent_user"] = getStringValue(parentMsg, "user")
	metadataMap["parent_ts"] = parentTS
	metadataMap["thread_ts"] = getStringValue(parentMsg, "thread_ts")
	metadataMap["team"] = getStringValue(parentMsg, "team")
	metadataMap["reply_count"] = getIntValue(parentMsg, "reply_count")
	metadataMap["reply_users_count"] = getIntValue(parentMsg, "reply_users_count")

	context.Metadata = metadataMap

	return EnrichResponse{
		Context: context,
		Status:  EnrichResponseStatusEnumSuccess,
	}
}

func parseSlackTS(ts string) (time.Time, error) {
	parts := strings.Split(ts, ".")

	secStr := parts[0]
	var nsec int64

	sec, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	if len(parts) > 1 {
		usecStr := parts[1]
		usec, err := strconv.ParseInt(usecStr, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		nsec = usec * 1000
	}

	return time.Unix(sec, nsec).UTC(), nil
}

// ============================================================================
// Test Export Functions
// ============================================================================

//go:export ApplyChannelEnrichment
func ApplyChannelEnrichment() int32 {
	input := pdk.Input()

	var apiResp map[string]any
	if err := json.Unmarshal(input, &apiResp); err != nil {
		pdk.SetError(err)
		return 1
	}

	gen := NewContextGenerator("test-workspace.slack.com")
	context := gen.CreateChannelContext("C099VUEKVBN", "general")

	response := applyChannelEnrichment(context, apiResp, "test-workspace.slack.com")

	output, err := json.Marshal(response.Context)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}

//go:export ApplyThreadEnrichment
func ApplyThreadEnrichment() int32 {
	input := pdk.Input()

	var apiResp map[string]any
	if err := json.Unmarshal(input, &apiResp); err != nil {
		pdk.SetError(err)
		return 1
	}

	gen := NewContextGenerator("test-workspace.slack.com")
	context := gen.CreateThreadContext("C099VUEKVBN", "1765613134.990399")

	response := applyThreadEnrichment(context, apiResp, "test-workspace.slack.com", "C099VUEKVBN")

	output, err := json.Marshal(response.Context)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}
