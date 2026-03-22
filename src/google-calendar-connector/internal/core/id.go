package core

import "fmt"

const (
	// ConnectorID is the unique identifier for this connector
	ConnectorID = "google-calendar"
	// CalendarAPIBase is the base URL for Google Calendar REST API v3
	CalendarAPIBase = "https://www.googleapis.com/calendar/v3"
)

// Resource type constants for context identification
const (
	ResourceTypeSource   = "source"
	ResourceTypeCalendar = "calendar"
	ResourceTypeEvent    = "event"
)

// MakeActivityID creates an activity ID for a calendar event
func MakeActivityID(calendarID, eventID string) string {
	return fmt.Sprintf("%s:%s:%s", ConnectorID, calendarID, eventID)
}

// MakeSourceContextID creates a source context ID
func MakeSourceContextID() string {
	return fmt.Sprintf("%s:%s", ConnectorID, ResourceTypeSource)
}

// MakeCalendarContextID creates a calendar context ID
func MakeCalendarContextID(calendarID string) string {
	return fmt.Sprintf("%s:%s:%s", ConnectorID, ResourceTypeCalendar, calendarID)
}

// MakeEventContextID creates an event context ID
func MakeEventContextID(calendarID, eventID string) string {
	return fmt.Sprintf("%s:%s:%s:%s", ConnectorID, ResourceTypeEvent, calendarID, eventID)
}
