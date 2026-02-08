package core

import "fmt"

const (
	// ConnectorID is the unique identifier for this connector
	ConnectorID = "slack"
	// SlackAPIBaseURL is the base URL for Slack API
	SlackAPIBaseURL = "https://slack.com/api"
)

// Resource type constants for context identification
const (
	ResourceTypeSource  = "source"
	ResourceTypeChannel = "channel"
	ResourceTypeThread  = "thread"
)

// MakeActivityID creates an activity ID with connector prefix
func MakeActivityID(messageTS string) string {
	return fmt.Sprintf("%s:%s", ConnectorID, messageTS)
}

// MakeSourceContextID creates a source context ID with connector prefix
func MakeSourceContextID() string {
	return fmt.Sprintf("%s:%s", ConnectorID, ResourceTypeSource)
}

// MakeChannelContextID creates a channel context ID with connector prefix
func MakeChannelContextID(channelID string) string {
	return fmt.Sprintf("%s:%s:%s", ConnectorID, ResourceTypeChannel, channelID)
}

// MakeThreadContextID creates a thread context ID with connector prefix
func MakeThreadContextID(channelID, threadTS string) string {
	return fmt.Sprintf("%s:%s:%s:%s", ConnectorID, ResourceTypeThread, channelID, threadTS)
}
