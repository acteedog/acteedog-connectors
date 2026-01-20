package main

import "fmt"

// makeActivityID creates an activity ID with connector prefix
func makeActivityID(messageTS string) string {
	return fmt.Sprintf("%s:%s", ConnectorID, messageTS)
}

// makeContextID creates a context ID with connector prefix
func makeContextID(contextType string, values ...string) string {
	base := fmt.Sprintf("%s:%s", ConnectorID, contextType)
	for _, v := range values {
		if v != "" {
			base = fmt.Sprintf("%s:%s", base, v)
		}
	}
	return base
}
