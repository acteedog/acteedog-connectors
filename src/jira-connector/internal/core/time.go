package core

import (
	"fmt"
	"time"
)

// jiraTimeFormats lists the time formats Jira may return.
// Jira API returns timestamps with a timezone offset without a colon (e.g. "+0900")
// which is not strictly RFC3339-compliant, so we try multiple formats.
var jiraTimeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999-0700",
	"2006-01-02T15:04:05-0700",
}

// ParseJiraTime parses a Jira timestamp string, trying multiple formats.
func ParseJiraTime(s string) (time.Time, error) {
	for _, f := range jiraTimeFormats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as a Jira timestamp", s)
}
