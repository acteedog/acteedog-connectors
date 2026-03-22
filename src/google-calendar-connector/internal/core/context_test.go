package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSourceContext(t *testing.T) {
	g := NewContextGenerator()
	ctx := g.CreateSourceContext()

	assert.Equal(t, "google-calendar:source", ctx.Id)
	assert.Equal(t, "", ctx.ParentId)
	assert.Equal(t, ConnectorID, ctx.ConnectorId)
	assert.Equal(t, ResourceTypeSource, ctx.ResourceType)
	assert.NotNil(t, ctx.Title)
	assert.Equal(t, "Google Calendar", *ctx.Title)

	meta, ok := ctx.Metadata.(map[string]any)
	assert.True(t, ok)
	_, hasEnrichmentParams := meta["enrichment_params"]
	assert.True(t, hasEnrichmentParams)
}

func TestCreateCalendarContext(t *testing.T) {
	g := NewContextGenerator()
	ctx := g.CreateCalendarContext("primary", "My Calendar")

	assert.Equal(t, "google-calendar:calendar:primary", ctx.Id)
	assert.Equal(t, "google-calendar:source", ctx.ParentId)
	assert.Equal(t, ConnectorID, ctx.ConnectorId)
	assert.Equal(t, ResourceTypeCalendar, ctx.ResourceType)
	assert.NotNil(t, ctx.Title)
	assert.Equal(t, "My Calendar", *ctx.Title)

	meta, ok := ctx.Metadata.(map[string]any)
	assert.True(t, ok)
	params, ok := meta["enrichment_params"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "primary", params["calendar_id"])
}

func TestCreateEventContext(t *testing.T) {
	g := NewContextGenerator()
	ctx := g.CreateEventContext("primary", "evt1", "Team Standup")

	assert.Equal(t, "google-calendar:event:primary:evt1", ctx.Id)
	assert.Equal(t, "google-calendar:calendar:primary", ctx.ParentId)
	assert.Equal(t, ConnectorID, ctx.ConnectorId)
	assert.Equal(t, ResourceTypeEvent, ctx.ResourceType)
	assert.NotNil(t, ctx.Title)
	assert.Equal(t, "Team Standup", *ctx.Title)

	meta, ok := ctx.Metadata.(map[string]any)
	assert.True(t, ok)
	params, ok := meta["enrichment_params"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "primary", params["calendar_id"])
	assert.Equal(t, "evt1", params["event_id"])
}
