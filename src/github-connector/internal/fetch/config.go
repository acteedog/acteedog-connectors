package fetch

import (
	"fmt"
	"time"
)

type config struct {
	token              string
	username           string
	repositoryPatterns []string
	startTime, endTime time.Time
}

func newConfig(cfg map[string]any, targetDate string) (*config, error) {
	token, ok := cfg["credential_personal_access_token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("missing personal access token")
	}

	username, ok := cfg["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("missing username")
	}

	var repositoryPatterns []string
	if patternsInterface, ok := cfg["repository_patterns"]; ok && patternsInterface != nil {
		if patterns, ok := patternsInterface.([]any); ok {
			for _, p := range patterns {
				if patternStr, ok := p.(string); ok && patternStr != "" {
					repositoryPatterns = append(repositoryPatterns, patternStr)
				}
			}
		}
	}

	startTime, endTime, err := parseDateRange(targetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target date: %w", err)
	}

	return &config{
		token:              token,
		username:           username,
		repositoryPatterns: repositoryPatterns,
		startTime:          startTime,
		endTime:            endTime,
	}, nil
}

func parseDateRange(targetDate string) (time.Time, time.Time, error) {
	var t time.Time
	var err error

	// Try RFC3339 format first (2025-12-12T00:00:00Z or 2025-12-12T00:00:00+00:00)
	t, err = time.Parse(time.RFC3339, targetDate)
	if err != nil {
		// Fallback to date-only format (2025-12-12)
		t, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	endTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)

	return startTime, endTime, nil
}
