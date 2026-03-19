package enrich

// CalendarDetailResponse is the response from GET /users/me/calendarList/{calendarId}
type CalendarDetailResponse struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	TimeZone    string `json:"timeZone"`
	AccessRole  string `json:"accessRole"`
}
