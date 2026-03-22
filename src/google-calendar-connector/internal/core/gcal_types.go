package core

// Event represents a Google Calendar event
type Event struct {
	ID             string          `json:"id"`
	Summary        string          `json:"summary"`
	Description    string          `json:"description"`
	Status         string          `json:"status"`
	HtmlLink       string          `json:"htmlLink"`
	Location       string          `json:"location"`
	Created        string          `json:"created"`
	Updated        string          `json:"updated"`
	Start          EventTime       `json:"start"`
	End            EventTime       `json:"end"`
	Attendees      []Attendee      `json:"attendees"`
	Creator        *Person         `json:"creator,omitempty"`
	Organizer      *Person         `json:"organizer,omitempty"`
	ConferenceData *ConferenceData `json:"conferenceData,omitempty"`
}

// EventTime represents the start or end time of an event
type EventTime struct {
	DateTime string `json:"dateTime"` // RFC3339, present for timed events
	Date     string `json:"date"`     // YYYY-MM-DD, present for all-day events
	TimeZone string `json:"timeZone"`
}

// Attendee represents an event attendee
type Attendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	ResponseStatus string `json:"responseStatus"`
	Self           bool   `json:"self"`
}

// Person represents an event creator or organizer
type Person struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Self        bool   `json:"self"`
}

// ConferenceData contains conference/video call information
type ConferenceData struct {
	EntryPoints []EntryPoint `json:"entryPoints"`
}

// EntryPoint represents a single way to join the conference
type EntryPoint struct {
	EntryPointType string `json:"entryPointType"`
	URI            string `json:"uri"`
}
