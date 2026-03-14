package fetch

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/core"
	"time"
)

type config struct {
	*core.ConnectorConfig
	startTime  time.Time
	endTime    time.Time
	targetDate string
}

func newConfig(cfg any, targetDate string) (*config, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var connCfg core.ConnectorConfig
	if err := json.Unmarshal(b, &connCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := connCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid connector config: %w", err)
	}

	startTime, endTime, parsedDate, err := parseDateRange(targetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target date: %w", err)
	}

	return &config{
		ConnectorConfig: &connCfg,
		startTime:       startTime,
		endTime:         endTime,
		targetDate:      parsedDate,
	}, nil
}

// parseDateRange parses the target date and returns start/end times and the date string (YYYY-MM-DD)
func parseDateRange(targetDate string) (time.Time, time.Time, string, error) {
	var t time.Time
	var err error

	// Try RFC3339 format first
	t, err = time.Parse(time.RFC3339, targetDate)
	if err != nil {
		// Fallback to date-only format
		t, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return time.Time{}, time.Time{}, "", err
		}
	}

	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	endTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
	dateStr := t.Format("2006-01-02")

	return startTime, endTime, dateStr, nil
}
