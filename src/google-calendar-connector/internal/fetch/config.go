package fetch

import (
	"time"
)

type config struct {
	startTime   time.Time
	endTime     time.Time
	targetEmail string
}

func newConfig(targetDate, targetEmail string) (*config, error) {
	startTime, endTime, err := parseDateRange(targetDate)
	if err != nil {
		return nil, err
	}

	return &config{
		startTime:   startTime,
		endTime:     endTime,
		targetEmail: targetEmail,
	}, nil
}

// parseDateRange parses the target date and returns start/end times in UTC
func parseDateRange(targetDate string) (time.Time, time.Time, error) {
	var t time.Time
	var err error

	// Try RFC3339 format first
	t, err = time.Parse(time.RFC3339, targetDate)
	if err != nil {
		// Fallback to date-only format
		t, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	endTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)

	return startTime, endTime, nil
}
