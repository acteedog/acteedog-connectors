package enrich

import (
	"encoding/json"
	"fmt"
	"google-calendar-connector/internal/core"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient is a simple in-test implementation of HTTPClient
type mockHTTPClient struct {
	calendarDetail    *CalendarDetailResponse
	calendarDetailErr error
	eventDetail       *core.Event
	eventDetailErr    error
}

func (m *mockHTTPClient) FetchCalendarDetail(calendarID string) (*CalendarDetailResponse, error) {
	return m.calendarDetail, m.calendarDetailErr
}

func (m *mockHTTPClient) FetchEventDetail(calendarID, eventID string) (*core.Event, error) {
	return m.eventDetail, m.eventDetailErr
}

func loadCalendarDetail(t *testing.T, path string) *CalendarDetailResponse {
	t.Helper()
	b, err := os.ReadFile(path) // nolint:gosec
	require.NoError(t, err)
	var data CalendarDetailResponse
	require.NoError(t, json.Unmarshal(b, &data))
	return &data
}

func loadEventDetail(t *testing.T, path string) *core.Event {
	t.Helper()
	b, err := os.ReadFile(path) // nolint:gosec
	require.NoError(t, err)
	var data core.Event
	require.NoError(t, json.Unmarshal(b, &data))
	return &data
}

func TestEnrichContext(t *testing.T) {
	calendarDetail := loadCalendarDetail(t, "../../testdata/enrichment/calendar_detail.json")
	eventDetail := loadEventDetail(t, "../../testdata/enrichment/event_detail.json")

	tests := []struct {
		name         string
		contextType  string
		params       any
		mock         *mockHTTPClient
		inputContext *core.Context
		checkFn      func(t *testing.T, ctx *core.Context)
		wantErr      bool
	}{
		{
			name:        "source: sets title, description, url",
			contextType: core.ResourceTypeSource,
			params:      map[string]any{},
			mock:        &mockHTTPClient{},
			inputContext: &core.Context{
				Id:           core.MakeSourceContextID(),
				ResourceType: core.ResourceTypeSource,
				ConnectorId:  core.ConnectorID,
				Metadata:     map[string]any{"enrichment_params": map[string]any{}},
			},
			checkFn: func(t *testing.T, ctx *core.Context) {
				require.NotNil(t, ctx.Title)
				assert.Equal(t, "Google Calendar", *ctx.Title)
				require.NotNil(t, ctx.Url)
				assert.Equal(t, core.CalendarAPIBase, *ctx.Url)
			},
		},
		{
			name:        "calendar: enriches with timezone, access_role",
			contextType: core.ResourceTypeCalendar,
			params: map[string]any{
				"calendar_id": "you@example.com",
			},
			mock: &mockHTTPClient{calendarDetail: calendarDetail},
			inputContext: &core.Context{
				Id:           core.MakeCalendarContextID("you@example.com"),
				ResourceType: core.ResourceTypeCalendar,
				ConnectorId:  core.ConnectorID,
				Metadata: map[string]any{
					"enrichment_params": map[string]any{"calendar_id": "you@example.com"},
				},
			},
			checkFn: func(t *testing.T, ctx *core.Context) {
				require.NotNil(t, ctx.Title)
				assert.Equal(t, "you@exmaple.com", *ctx.Title)
				meta, ok := ctx.Metadata.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "Asia/Tokyo", meta["timezone"])
				assert.Equal(t, "owner", meta["access_role"])
			},
		},
		{
			name:        "event: enriches with attendees, conference_link, organizer, creator",
			contextType: core.ResourceTypeEvent,
			params: map[string]any{
				"calendar_id": "you@example.com",
				"event_id":    "calendar-id-2",
			},
			mock: &mockHTTPClient{eventDetail: eventDetail},
			inputContext: &core.Context{
				Id:           core.MakeEventContextID("you@example.com", "calendar-id-2"),
				ResourceType: core.ResourceTypeEvent,
				ConnectorId:  core.ConnectorID,
				Metadata: map[string]any{
					"enrichment_params": map[string]any{
						"calendar_id": "you@example.com",
						"event_id":    "calendar-id-2",
					},
				},
			},
			checkFn: func(t *testing.T, ctx *core.Context) {
				require.NotNil(t, ctx.Title)
				assert.Equal(t, "Invited event 2", *ctx.Title)
				require.NotNil(t, ctx.CreatedAt)
				require.NotNil(t, ctx.UpdatedAt)

				meta, ok := ctx.Metadata.(map[string]any)
				require.True(t, ok)
				_, hasLocation := meta["location"]
				assert.False(t, hasLocation, "location should not be set when empty")
				assert.Equal(t, "another@example.com", meta["organizer_email"])
				assert.Equal(t, "another@example.com", meta["creator_email"])
				assert.Equal(t, "https://meet.google.com/meet-id", meta["conference_link"])
				attendees, ok := meta["attendees"].([]string)
				require.True(t, ok)
				assert.Len(t, attendees, 2)
			},
		},
		{
			name:        "unsupported context type returns error",
			contextType: "unknown",
			params:      map[string]any{},
			mock:        &mockHTTPClient{},
			inputContext: &core.Context{
				ResourceType: "unknown",
			},
			wantErr: true,
		},
		{
			name:        "calendar: API error returns error",
			contextType: core.ResourceTypeCalendar,
			params: map[string]any{
				"calendar_id": "primary",
			},
			mock: &mockHTTPClient{calendarDetailErr: fmt.Errorf("API error")},
			inputContext: &core.Context{
				ResourceType: core.ResourceTypeCalendar,
				Metadata:     map[string]any{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enricher, err := NewContextEnricher(tt.mock, tt.contextType, tt.params, &core.NoopLogger{})
			require.NoError(t, err)

			result, err := enricher.EnrichContext(tt.inputContext)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.checkFn(t, result)
		})
	}
}
