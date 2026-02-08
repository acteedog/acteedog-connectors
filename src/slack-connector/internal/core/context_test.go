package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSourceContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreateSourceContext()
	want := &Context{
		Id:           "slack:source",
		Name:         "slack:source",
		Level:        1,
		ParentId:     "",
		ConnectorId:  "slack",
		ResourceType: "source",
		Title:        ptrString("Slack"),
		Description:  ptrString("Activity source from Slack"),
		Url:          ptrString("https://slack.com"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateChannelContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreateChannelContext("C1234567890", "general")
	want := &Context{
		Id:           "slack:channel:C1234567890",
		Name:         "channel #general",
		Level:        2,
		ParentId:     "slack:source",
		ConnectorId:  "slack",
		ResourceType: "channel",
		Title:        ptrString("#general"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"channel_id": "C1234567890",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestCreateThreadContext(t *testing.T) {
	g := NewContextGenerator()
	got := g.CreateThreadContext("C1234567890", "1623855600.000200")
	want := &Context{
		Id:           "slack:thread:C1234567890:1623855600.000200",
		Name:         "Thread 1623855600.000200",
		Level:        3,
		ParentId:     "slack:channel:C1234567890",
		ConnectorId:  "slack",
		ResourceType: "thread",
		Title:        ptrString("Thread 1623855600.000200"),
		Metadata: map[string]any{
			"enrichment_params": map[string]any{
				"channel_id": "C1234567890",
				"thread_ts":  "1623855600.000200",
			},
		},
	}
	assert.Equal(t, want, got)
}
