package fetch

import (
	"encoding/json"
	"fmt"
	"google-calendar-connector/internal/core"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const targetEmail = "you@example.com"

// mockHTTPClient is a simple in-test implementation of HTTPClient
type mockHTTPClient struct {
	calendarList    *CalendarListResponse
	calendarListErr error
	eventResponses  map[string]*EventListResponse
	eventErrors     map[string]error
}

func (m *mockHTTPClient) FetchCalendarList() (*CalendarListResponse, error) {
	return m.calendarList, m.calendarListErr
}

func (m *mockHTTPClient) FetchEvents(calendarID, timeMin, timeMax string) (*EventListResponse, error) {
	if m.eventErrors != nil {
		if err, ok := m.eventErrors[calendarID]; ok {
			return nil, err
		}
	}
	if m.eventResponses != nil {
		if resp, ok := m.eventResponses[calendarID]; ok {
			return resp, nil
		}
	}
	return &EventListResponse{}, nil
}

func loadCalendarList(t *testing.T, path string) *CalendarListResponse {
	t.Helper()
	b, err := os.ReadFile(path) // nolint:gosec
	require.NoError(t, err)
	var data CalendarListResponse
	require.NoError(t, json.Unmarshal(b, &data))
	return &data
}

func loadEventList(t *testing.T, path string) *EventListResponse {
	t.Helper()
	b, err := os.ReadFile(path) // nolint:gosec
	require.NoError(t, err)
	var data EventListResponse
	require.NoError(t, json.Unmarshal(b, &data))
	return &data
}

func mergeEventLists(lists ...*EventListResponse) *EventListResponse {
	merged := &EventListResponse{}
	for _, l := range lists {
		merged.Items = append(merged.Items, l.Items...)
	}
	return merged
}

func TestFetchActivities(t *testing.T) {
	calendarList := loadCalendarList(t, "../../testdata/fetch/calendar_list.json")
	eventsOwned := loadEventList(t, "../../testdata/fetch/events_owned.json")
	eventsInvited := loadEventList(t, "../../testdata/fetch/events_invited.json")

	tests := []struct {
		name            string
		calendarList    *CalendarListResponse
		calendarListErr error
		eventResponses  map[string]*EventListResponse
		targetEmail     string
		targetDate      string
		wantCount       int
		wantIDs         []string
		wantErr         bool
	}{
		{
			name:         "normal: owned events included, accepted invited included, needsAction excluded, other calendar excluded",
			calendarList: calendarList,
			eventResponses: map[string]*EventListResponse{
				targetEmail: mergeEventLists(eventsOwned, eventsInvited),
			},
			targetEmail: targetEmail,
			targetDate:  "2026-03-15",
			// owned: calendar-id-1, calendar-id-2 (both creator=you)
			// invited: calendar-id-1 (needsAction→excluded), calendar-id-2 (accepted→included)
			wantCount: 3,
			wantIDs: []string{
				"google-calendar:" + targetEmail + ":calendar-id-1",
				"google-calendar:" + targetEmail + ":calendar-id-2",
				"google-calendar:" + targetEmail + ":calendar-id-2",
			},
		},
		{
			name:         "other calendar (test.calendar@example.com) is filtered out",
			calendarList: calendarList,
			eventResponses: map[string]*EventListResponse{
				targetEmail:                 eventsOwned,
				"test.calendar@example.com": eventsInvited,
			},
			targetEmail: targetEmail,
			targetDate:  "2026-03-15",
			wantCount:   2,
		},
		{
			name:         "needsAction attendee is excluded",
			calendarList: calendarList,
			eventResponses: map[string]*EventListResponse{
				targetEmail: eventsInvited,
			},
			targetEmail: targetEmail,
			targetDate:  "2026-03-16",
			// calendar-id-1: needsAction → excluded
			// calendar-id-2: accepted → included
			wantCount: 1,
			wantIDs:   []string{"google-calendar:" + targetEmail + ":calendar-id-2"},
		},
		{
			name:            "calendar list API error returns error",
			calendarListErr: fmt.Errorf("API error"),
			targetEmail:     targetEmail,
			targetDate:      "2026-03-15",
			wantErr:         true,
		},
		{
			name: "deleted calendar skipped",
			calendarList: &CalendarListResponse{
				Items: []CalendarListEntry{
					{ID: "deleted@example.com", Summary: "Deleted", Deleted: true},
				},
			},
			targetEmail: "deleted@example.com",
			targetDate:  "2026-03-15",
			wantCount:   0,
		},
		{
			name: "all-day event has UTC midnight timestamp",
			calendarList: &CalendarListResponse{
				Items: []CalendarListEntry{
					{ID: targetEmail, Summary: "My Calendar"},
				},
			},
			eventResponses: map[string]*EventListResponse{
				targetEmail: {
					Items: []core.Event{
						{
							ID:        "allday1",
							Summary:   "All Day Event",
							Status:    "confirmed",
							Start:     core.EventTime{Date: "2026-03-15"},
							End:       core.EventTime{Date: "2026-03-16"},
							Organizer: &core.Person{Email: targetEmail},
						},
					},
				},
			},
			targetEmail: targetEmail,
			targetDate:  "2026-03-15",
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockHTTPClient{
				calendarList:    tt.calendarList,
				calendarListErr: tt.calendarListErr,
				eventResponses:  tt.eventResponses,
			}

			fetcher, err := NewActivityFetcher(mock, tt.targetDate, tt.targetEmail, &core.NoopLogger{})
			require.NoError(t, err)

			activities, err := fetcher.FetchActivities()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, activities, tt.wantCount)

			if len(tt.wantIDs) > 0 {
				gotIDs := make([]string, len(activities))
				for i, a := range activities {
					gotIDs[i] = a.Id
				}
				assert.Equal(t, tt.wantIDs, gotIDs)
			}
		})
	}
}

func TestFetchActivities_AllDayTimestamp(t *testing.T) {
	mock := &mockHTTPClient{
		calendarList: &CalendarListResponse{
			Items: []CalendarListEntry{
				{ID: targetEmail, Summary: "My Calendar"},
			},
		},
		eventResponses: map[string]*EventListResponse{
			targetEmail: {
				Items: []core.Event{
					{
						ID:        "allday1",
						Summary:   "All Day",
						Status:    "confirmed",
						Start:     core.EventTime{Date: "2026-03-15"},
						End:       core.EventTime{Date: "2026-03-16"},
						Organizer: &core.Person{Email: targetEmail},
					},
				},
			},
		},
	}

	fetcher, err := NewActivityFetcher(mock, "2026-03-15", targetEmail, &core.NoopLogger{})
	require.NoError(t, err)

	activities, err := fetcher.FetchActivities()
	require.NoError(t, err)
	require.Len(t, activities, 1)

	expected := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, activities[0].Timestamp)

	meta, ok := activities[0].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, meta["is_all_day"])
	assert.Equal(t, "2026-03-15", meta["start_time"])
	assert.Equal(t, "2026-03-16", meta["end_time"])
}

func TestFetchActivities_Contexts(t *testing.T) {
	mock := &mockHTTPClient{
		calendarList: &CalendarListResponse{
			Items: []CalendarListEntry{
				{ID: targetEmail, Summary: "My Calendar"},
			},
		},
		eventResponses: map[string]*EventListResponse{
			targetEmail: {
				Items: []core.Event{
					{
						ID:        "evt1",
						Summary:   "Meeting",
						Status:    "confirmed",
						Start:     core.EventTime{DateTime: "2026-03-15T10:00:00Z"},
						End:       core.EventTime{DateTime: "2026-03-15T11:00:00Z"},
						Organizer: &core.Person{Email: targetEmail},
					},
				},
			},
		},
	}

	fetcher, err := NewActivityFetcher(mock, "2026-03-15", targetEmail, &core.NoopLogger{})
	require.NoError(t, err)

	activities, err := fetcher.FetchActivities()
	require.NoError(t, err)
	require.Len(t, activities, 1)

	contexts := activities[0].Contexts
	require.Len(t, contexts, 3)
	assert.Equal(t, core.ResourceTypeSource, contexts[0].ResourceType)
	assert.Equal(t, core.ResourceTypeCalendar, contexts[1].ResourceType)
	assert.Equal(t, core.ResourceTypeEvent, contexts[2].ResourceType)
	assert.Equal(t, "google-calendar:source", contexts[0].Id)
	assert.Equal(t, "google-calendar:calendar:"+targetEmail, contexts[1].Id)
	assert.Equal(t, "google-calendar:event:"+targetEmail+":evt1", contexts[2].Id)
}

func TestFetchActivities_Metadata(t *testing.T) {
	eventsOwned := loadEventList(t, "../../testdata/fetch/events_owned.json")
	eventsInvited := loadEventList(t, "../../testdata/fetch/events_invited.json")

	mock := &mockHTTPClient{
		calendarList: &CalendarListResponse{
			Items: []CalendarListEntry{
				{ID: targetEmail, Summary: "My Calendar"},
			},
		},
		eventResponses: map[string]*EventListResponse{
			targetEmail: mergeEventLists(eventsOwned, eventsInvited),
		},
	}

	fetcher, err := NewActivityFetcher(mock, "2026-03-15", targetEmail, &core.NoopLogger{})
	require.NoError(t, err)

	activities, err := fetcher.FetchActivities()
	require.NoError(t, err)
	require.Len(t, activities, 3)

	// First owned event: creator_email should be set, no my_response_status (not an attendee)
	ownedMeta, ok := activities[0].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, targetEmail, ownedMeta["creator_email"])
	assert.Equal(t, "2026-03-15T10:00:00+09:00", ownedMeta["start_time"])
	assert.Equal(t, "2026-03-15T12:00:00+09:00", ownedMeta["end_time"])
	_, hasMyStatus := ownedMeta["my_response_status"]
	assert.False(t, hasMyStatus, "owned event should not have my_response_status (user is not an attendee)")

	// Invited accepted event (calendar-id-2 from events_invited): my_response_status = "accepted"
	invitedMeta, ok := activities[2].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "another@example.com", invitedMeta["creator_email"])
	assert.Equal(t, "accepted", invitedMeta["my_response_status"])
}
