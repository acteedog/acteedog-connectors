package enrich

import (
	"fmt"
	"slack-connector/internal/core"
	"strconv"
	"strings"
	"time"
)

type ContextEnricher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewContextEnricher creates a new ContextEnricher instance
func NewContextEnricher(httpClient HTTPClient, contextType string, cfg, params map[string]any, logger core.Logger) (*ContextEnricher, error) {
	config, err := newConfig(contextType, cfg, params)
	if err != nil {
		return nil, err
	}

	return &ContextEnricher{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}, nil
}

// EnrichContext enriches the given context with additional data from GitHub
func (e *ContextEnricher) EnrichContext(context *core.Context) (*core.Context, error) {
	e.logger.Info("Starting to enrich context")

	switch e.config.contextType {
	case core.ResourceTypeSource:
		context.Title = ptrString("Slack")
		context.Description = ptrString("Activity source from Slack")
		context.Url = ptrString("https://" + e.config.workspaceURL)

		return context, nil
	case core.ResourceTypeChannel:
		return e.enrichChannel(context)
	case core.ResourceTypeThread:
		return e.enrichThread(context)
	default:
		return nil, fmt.Errorf("unsupported context type: %s", e.config.contextType)
	}
}

func (e *ContextEnricher) enrichChannel(context *core.Context) (*core.Context, error) {
	channelID, ok := e.config.enrichmentParams["channel_id"].(string)
	if !ok || channelID == "" {
		return nil, fmt.Errorf("channel_id not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching channel: %s", channelID))

	response, err := e.httpClient.FetchChannel(e.config.token, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel data: %w", err)
	}

	return e.applyChannelEnrichment(context, response)
}

func (e *ContextEnricher) applyChannelEnrichment(context *core.Context, apiResp map[string]any) (*core.Context, error) {
	channelObj, ok := apiResp["channel"].(map[string]any)
	if !ok {
		return context, fmt.Errorf("invalid channel data in API response")
	}

	name := core.GetStringValue(channelObj, "name")
	topicValue := getNestedString(channelObj, "topic", "value")
	channelID := core.GetStringValue(channelObj, "id")

	title := fmt.Sprintf("#%s", name)
	description := topicValue
	url := fmt.Sprintf("https://%s/archives/%s", e.config.workspaceURL, channelID)
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
	metadataMap["context_team_id"] = core.GetStringValue(channelObj, "context_team_id")

	context.Metadata = metadataMap

	return context, nil
}

func (e *ContextEnricher) enrichThread(context *core.Context) (*core.Context, error) {
	channelID, ok := e.config.enrichmentParams["channel_id"].(string)
	if !ok || channelID == "" {
		return nil, fmt.Errorf("channel_id not found in enrichment_params")
	}

	threadTS, ok := e.config.enrichmentParams["thread_ts"].(string)
	if !ok || threadTS == "" {
		return nil, fmt.Errorf("thread_ts not found in enrichment_params")
	}

	e.logger.Info(fmt.Sprintf("Enriching thread: %s in channel %s", threadTS, channelID))

	response, err := e.httpClient.FetchThread(e.config.token, channelID, threadTS)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread data: %w", err)
	}

	return e.applyThreadEnrichment(context, response, channelID)
}

func (e *ContextEnricher) applyThreadEnrichment(context *core.Context, apiResp map[string]any, channelID string) (*core.Context, error) {
	messagesInterface, ok := apiResp["messages"]
	if !ok {
		return nil, fmt.Errorf("no messages array in API response")
	}

	messages, ok := messagesInterface.([]any)
	if !ok || len(messages) == 0 {
		return nil, fmt.Errorf("messages array is empty or invalid")
	}

	// Get first message (parent message)
	parentMsg, ok := messages[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("first message is not a map")
	}

	text := core.GetStringValue(parentMsg, "text")
	parentTS := core.GetStringValue(parentMsg, "ts")

	title := fmt.Sprintf("Thread: %s", text)
	description := text
	// Format: https://{workspace_url}/archives/{channel_id}/p{ts with dots removed}
	url := fmt.Sprintf("https://%s/archives/%s/p%s", e.config.workspaceURL, channelID, formatSlackTS(parentTS))
	createdAt, err := parseSlackTS(core.GetStringValue(parentMsg, "thread_ts"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse createdAt for thread: %w", err)
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

	metadataMap["parent_user"] = core.GetStringValue(parentMsg, "user")
	metadataMap["parent_ts"] = parentTS
	metadataMap["thread_ts"] = core.GetStringValue(parentMsg, "thread_ts")
	metadataMap["team"] = core.GetStringValue(parentMsg, "team")
	metadataMap["reply_count"] = getIntValue(parentMsg, "reply_count")
	metadataMap["reply_users_count"] = getIntValue(parentMsg, "reply_users_count")

	context.Metadata = metadataMap

	return context, nil
}

// getNestedString safely extracts nested string value
func getNestedString(m map[string]any, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - extract string
			return core.GetStringValue(current, key)
		}
		// Navigate deeper
		if nested, ok := current[key].(map[string]any); ok {
			current = nested
		} else {
			return ""
		}
	}
	return ""
}

// getIntValue safely extracts int64 value from map
func getIntValue(m map[string]any, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	return 0
}

// getBoolValue safely extracts bool value from map
func getBoolValue(m map[string]any, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}

// formatSlackTS converts Slack timestamp to URL format
// Example: "1765611321.248519" -> "1765611321248519"
func formatSlackTS(ts string) string {
	return strings.ReplaceAll(ts, ".", "")
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
