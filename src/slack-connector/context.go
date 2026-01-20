package main

import "fmt"

// ContextGenerator provides factory methods for creating standardized Context objects
type ContextGenerator struct {
	connectorID  string
	workspaceURL string
}

// NewContextGenerator creates a new ContextGenerator
func NewContextGenerator(workspaceURL string) *ContextGenerator {
	return &ContextGenerator{
		connectorID:  ConnectorID,
		workspaceURL: workspaceURL,
	}
}

// CreateSourceContext creates a Level 1 source context for Slack
func (g *ContextGenerator) CreateSourceContext() Context {
	id := makeContextID(ResourceTypeSource)
	return Context{
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
func (g *ContextGenerator) CreateChannelContext(channelID, channelName string) Context {
	id := makeContextID(ResourceTypeChannel, channelID)
	parentID := makeContextID(ResourceTypeSource)
	return Context{
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
func (g *ContextGenerator) CreateThreadContext(channelID, threadTS string) Context {
	id := makeContextID(ResourceTypeThread, channelID, threadTS)
	parentID := makeContextID(ResourceTypeChannel, channelID)
	return Context{
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
