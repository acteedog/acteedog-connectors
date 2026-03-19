package enrich

import "google-calendar-connector/internal/core"

// HTTPClient defines the interface for fetching data from the Google Calendar API for enrichment
type HTTPClient interface {
	// FetchCalendarDetail fetches details for a specific calendar
	FetchCalendarDetail(calendarID string) (*CalendarDetailResponse, error)
	// FetchEventDetail fetches details for a specific event
	FetchEventDetail(calendarID, eventID string) (*core.Event, error)
}
