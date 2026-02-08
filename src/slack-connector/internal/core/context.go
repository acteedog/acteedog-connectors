package core

import (
	"fmt"
	"time"
)

type Context struct {
	ConnectorId  string
	CreatedAt    *time.Time
	Description  *string
	Id           string
	Level        int64
	Metadata     any
	Name         string
	ParentId     string
	ResourceType string
	Title        *string
	UpdatedAt    *time.Time
	Url          *string
}

// ContextGenerator provides factory methods for creating standardized Context objects
type ContextGenerator struct {
	connectorID string
}

// NewContextGenerator creates a new ContextGenerator
func NewContextGenerator() *ContextGenerator {
	return &ContextGenerator{
		connectorID: ConnectorID,
	}
}

// CreateSourceContext creates a Level 1 source context for Slack
func (g *ContextGenerator) CreateSourceContext() *Context {
	id := MakeSourceContextID()
	return &Context{
		Id:           id,
		Name:         id,
		Level:        1,
		ParentId:     "", // Top level - no parent
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeSource,
		Title:        ptrString("Slack"),
		Description:  ptrString("Activity source from Slack"),
		Url:          ptrString("https://slack.com"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
}

// CreateChannelContext creates a Level 2 channel context
func (g *ContextGenerator) CreateChannelContext(channelID, channelName string) *Context {
	id := MakeChannelContextID(channelID)
	parentID := MakeSourceContextID()
	return &Context{
		Id:           id,
		Name:         fmt.Sprintf("channel #%s", channelName),
		Level:        2,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeChannel,
		Title:        ptrString(fmt.Sprintf("#%s", channelName)),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"channel_id": channelID,
			},
		},
	}
}

// CreateThreadContext creates a Level 3 thread context
func (g *ContextGenerator) CreateThreadContext(channelID, threadTS string) *Context {
	id := MakeThreadContextID(channelID, threadTS)
	parentID := MakeChannelContextID(channelID)
	return &Context{
		Id:           id,
		Name:         fmt.Sprintf("Thread %s", threadTS),
		Level:        3,
		ParentId:     parentID,
		ConnectorId:  g.connectorID,
		ResourceType: ResourceTypeThread,
		Title:        ptrString(fmt.Sprintf("Thread %s", threadTS)),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"channel_id": channelID,
				"thread_ts":  threadTS,
			},
		},
	}
}

// GetStringValue safely extracts string value from map
func GetStringValue(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}
