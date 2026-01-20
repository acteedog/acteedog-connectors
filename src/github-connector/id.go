package main

import "fmt"

// makeActivityID creates an activity ID with connector prefix
func makeActivityID(eventID string) string {
	return fmt.Sprintf("%s:%s", ConnectorID, eventID)
}

// makeContextID creates a context ID with connector prefix
func makeContextID(contextType string, value string) string {
	if value == "" {
		return fmt.Sprintf("%s:%s", ConnectorID, contextType)
	}
	return fmt.Sprintf("%s:%s:%s", ConnectorID, contextType, value)
}
