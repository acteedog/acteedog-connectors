package fetch

import "google-calendar-connector/internal/core"

// CalendarListResponse is the response from GET /users/me/calendarList
type CalendarListResponse struct {
	Items []CalendarListEntry `json:"items"`
}

// CalendarListEntry represents a single calendar in the list
type CalendarListEntry struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	TimeZone    string `json:"timeZone"`
	AccessRole  string `json:"accessRole"`
	Deleted     bool   `json:"deleted"`
}

// EventListResponse is the response from GET /calendars/{calId}/events
type EventListResponse struct {
	Items []core.Event `json:"items"`
}
