package fetch

import (
	"fmt"
	"google-calendar-connector/internal/core"
	"time"
	"unicode/utf8"
)

const (
	maxDescriptionLength = 500
)

// Activity represents a fetched calendar event activity
type Activity struct {
	ActivityType string
	Contexts     []*core.Context
	Description  string
	Id           string
	Metadata     any
	Source       string
	Timestamp    time.Time
	Title        string
	Url          *string
}

// ActivityFetcher fetches activities from Google Calendar
type ActivityFetcher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewActivityFetcher creates a new ActivityFetcher
func NewActivityFetcher(httpClient HTTPClient, targetDate, targetEmail string, logger core.Logger) (*ActivityFetcher, error) {
	cfg, err := newConfig(targetDate, targetEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid target date: %w", err)
	}

	return &ActivityFetcher{
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}, nil
}

// FetchActivities fetches calendar events and returns them as activities
func (f *ActivityFetcher) FetchActivities() ([]*Activity, error) {
	f.logger.Info("Fetching calendar list")

	calList, err := f.httpClient.FetchCalendarList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar list: %w", err)
	}

	cgen := core.NewContextGenerator()
	sourceCtx := cgen.CreateSourceContext()

	timeMin := f.config.startTime.Format(time.RFC3339)
	timeMax := f.config.endTime.Format(time.RFC3339)

	var activities []*Activity

	for i := range calList.Items {
		cal := &calList.Items[i]
		if cal.Deleted {
			continue
		}
		if cal.ID != f.config.targetEmail {
			continue
		}

		f.logger.Info(fmt.Sprintf("Fetching events for calendar: %s", cal.ID))

		calCtx := cgen.CreateCalendarContext(cal.ID, cal.Summary)

		events, err := f.httpClient.FetchEvents(cal.ID, timeMin, timeMax)
		if err != nil {
			f.logger.Warn(fmt.Sprintf("Failed to fetch events for calendar %s: %v", cal.ID, err))
			continue
		}

		for j := range events.Items {
			evt := &events.Items[j]
			if evt.Status == "cancelled" {
				continue
			}
			if !isEventForUser(evt, f.config.targetEmail) {
				continue
			}

			ts, isAllDay, err := parseEventTime(evt.Start)
			if err != nil {
				f.logger.Warn(fmt.Sprintf("Skipping event %s: invalid start time: %v", evt.ID, err))
				continue
			}

			eventCtx := cgen.CreateEventContext(cal.ID, evt.ID, evt.Summary)

			desc := evt.Description
			if utf8.RuneCountInString(desc) > maxDescriptionLength {
				runes := []rune(desc)
				desc = string(runes[:maxDescriptionLength])
			}

			var urlPtr *string
			if evt.HtmlLink != "" {
				u := evt.HtmlLink
				urlPtr = &u
			}

			attendeeCount := len(evt.Attendees)
			hasConference := evt.ConferenceData != nil && len(evt.ConferenceData.EntryPoints) > 0

			startTime := evt.Start.DateTime
			if startTime == "" {
				startTime = evt.Start.Date
			}
			endTime := evt.End.DateTime
			if endTime == "" {
				endTime = evt.End.Date
			}

			metadata := map[string]any{
				"status":         evt.Status,
				"is_all_day":     isAllDay,
				"attendee_count": attendeeCount,
				"has_conference": hasConference,
				"start_time":     startTime,
				"end_time":       endTime,
			}
			if evt.Creator != nil && evt.Creator.Email != "" {
				metadata["creator_email"] = evt.Creator.Email
			}
			for _, a := range evt.Attendees {
				if a.Email == f.config.targetEmail {
					metadata["my_response_status"] = a.ResponseStatus
					break
				}
			}

			activity := &Activity{
				Id:           core.MakeActivityID(cal.ID, evt.ID),
				Timestamp:    ts,
				Source:       core.ConnectorID,
				ActivityType: "calendar_event",
				Title:        evt.Summary,
				Description:  desc,
				Url:          urlPtr,
				Contexts:     []*core.Context{sourceCtx, calCtx, eventCtx},
				Metadata:     metadata,
			}
			activities = append(activities, activity)
		}
	}

	f.logger.Info(fmt.Sprintf("Fetched %d activities", len(activities)))

	return activities, nil
}

// isEventForUser returns true if the event is relevant to the given email:
// creator or organizer matches, or the user is an accepted attendee.
func isEventForUser(evt *core.Event, email string) bool {
	if evt.Creator != nil && evt.Creator.Email == email {
		return true
	}
	if evt.Organizer != nil && evt.Organizer.Email == email {
		return true
	}
	for _, a := range evt.Attendees {
		if a.Email == email && a.ResponseStatus == "accepted" {
			return true
		}
	}
	return false
}

// parseEventTime parses the event start time, returning the time.Time, whether it's all-day, and any error.
func parseEventTime(et core.EventTime) (time.Time, bool, error) {
	if et.DateTime != "" {
		// Timed event: RFC3339
		t, err := time.Parse(time.RFC3339, et.DateTime)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("failed to parse dateTime %q: %w", et.DateTime, err)
		}
		return t.UTC(), false, nil
	}
	if et.Date != "" {
		// All-day event: YYYY-MM-DD → UTC midnight
		t, err := time.Parse("2006-01-02", et.Date)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("failed to parse date %q: %w", et.Date, err)
		}
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true, nil
	}
	return time.Time{}, false, fmt.Errorf("event has neither dateTime nor date")
}

