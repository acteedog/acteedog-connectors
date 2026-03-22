package fetch

// HTTPClient defines the interface for fetching data from the Google Calendar API
type HTTPClient interface {
	// FetchCalendarList fetches the list of calendars for the authenticated user
	FetchCalendarList() (*CalendarListResponse, error)
	// FetchEvents fetches events from a specific calendar within the given time range
	FetchEvents(calendarID, timeMin, timeMax string) (*EventListResponse, error)
}
