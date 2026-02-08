package fetch

import (
	"fmt"
	"regexp"
	"slack-connector/internal/core"
	"strconv"
	"time"
)

type Activity struct {
	ActivityType string
	Contexts     []*core.Context
	Description  *string
	Id           string
	Metadata     any
	Source       string
	Timestamp    string
	Title        string
	Url          *string
}

// ActivityFetcher defines the structure for fetching activities from Slack
type ActivityFetcher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewActivityFetcher creates a new ActivityFetcher instance
func NewActivityFetcher(httpClient HTTPClient, cfg map[string]any, targetDate string, logger core.Logger) (*ActivityFetcher, error) {
	config, err := newConfig(cfg, targetDate)
	if err != nil {
		return nil, err
	}

	return &ActivityFetcher{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}, nil
}

// FetchActivities fetches and processes activities from Slack
func (f *ActivityFetcher) FetchActivities() ([]*Activity, error) {
	f.logger.Info("Starting to fetch Slack messages")

	allMessages, err := f.fetchAllMessages()
	if err != nil {
		return nil, err
	}

	f.logger.Info(fmt.Sprintf("Fetched %d messages", len(allMessages)))

	gen := core.NewContextGenerator()
	activities := []*Activity{}
	for _, message := range allMessages {
		activity, err := transformMessage(message, gen)
		if err != nil {
			f.logger.Warn(fmt.Sprintf("Skipping message: %s", err.Error()))
			continue
		}
		if activity != nil {
			activities = append(activities, activity)
		}
	}

	f.logger.Info(fmt.Sprintf("Transformed %d activities", len(activities)))

	return activities, nil
}

func (f *ActivityFetcher) fetchAllMessages() ([]map[string]any, error) {
	allMessages := []map[string]any{}

	// Slack search.messages API pagination (max 100 pages)
	for page := 1; page <= 100; page++ {
		f.logger.Debug(fmt.Sprintf("Fetching page %d", page))

		response, err := f.httpClient.FetchMessages(f.config.token, f.config.userID, f.config.targetDate, page)
		if err != nil {
			return nil, err
		}

		messagesObj, ok := response["messages"].(map[string]any)
		if !ok {
			f.logger.Debug("No messages object in response, stopping pagination")
			break
		}

		matchesInterface, ok := messagesObj["matches"]
		if !ok {
			f.logger.Debug("No matches in messages object, stopping pagination")
			break
		}

		matches, ok := matchesInterface.([]any)
		if !ok {
			f.logger.Debug("Matches is not an array, stopping pagination")
			break
		}

		if len(matches) == 0 {
			f.logger.Debug("Empty matches array, stopping pagination")
			break
		}

		for _, match := range matches {
			if msg, ok := match.(map[string]any); ok {
				allMessages = append(allMessages, msg)
			}
		}

		f.logger.Info(fmt.Sprintf("Page %d: %d messages fetched", page, len(matches)))

		paging, ok := messagesObj["paging"].(map[string]any)
		if !ok {
			break
		}

		pages, _ := paging["pages"].(float64)
		currentPage, _ := paging["page"].(float64)

		if currentPage >= pages {
			f.logger.Debug("Reached last page, stopping pagination")
			break
		}
	}

	return allMessages, nil
}

func transformMessage(message map[string]any, cgen *core.ContextGenerator) (*Activity, error) {
	ts := core.GetStringValue(message, "ts")
	if ts == "" {
		return nil, fmt.Errorf("message missing ts field")
	}

	text := core.GetStringValue(message, "text")
	permalink := core.GetStringValue(message, "permalink")
	username := core.GetStringValue(message, "username")
	team := core.GetStringValue(message, "team")

	channelObj, ok := message["channel"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("message missing channel object")
	}

	channelID := core.GetStringValue(channelObj, "id")
	channelName := core.GetStringValue(channelObj, "name")

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

	sourceContext := cgen.CreateSourceContext()
	channelContext := cgen.CreateChannelContext(channelID, channelName)
	threadContext := cgen.CreateThreadContext(channelID, threadTS)

	title := fmt.Sprintf("Message in #%s", channelName)
	description := text

	activity := Activity{
		Id:           core.MakeActivityID(ts),
		Timestamp:    timestamp.Format(time.RFC3339),
		Source:       core.ConnectorID,
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
		Contexts: []*core.Context{
			sourceContext,
			channelContext,
			threadContext,
		},
	}

	return &activity, nil
}

// parseThreadTS extracts thread_ts from permalink
// Example: https://...slack.com/archives/C099VUEKVBN/p1765613227980829?thread_ts=1765613134.990399
// Returns: "1765613134.990399" or empty string if not found
func parseThreadTS(permalink string) string {
	re := regexp.MustCompile(`thread_ts=([0-9.]+)`)
	matches := re.FindStringSubmatch(permalink)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// convertSlackTSToTime converts Slack timestamp to time.Time
// Slack timestamp format: "1765611321.248519" (epoch seconds with microseconds)
func convertSlackTSToTime(ts string) (time.Time, error) {
	// Parse as float to handle the decimal part
	epochFloat, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Convert to Unix timestamp (seconds and nanoseconds)
	seconds := int64(epochFloat)
	nanos := int64((epochFloat - float64(seconds)) * 1e9)

	return time.Unix(seconds, nanos).UTC(), nil
}
