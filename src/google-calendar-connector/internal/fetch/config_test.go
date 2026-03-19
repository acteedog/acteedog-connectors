package fetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		name        string
		targetDate  string
		wantStart   time.Time
		wantEnd     time.Time
		wantErr     bool
	}{
		{
			name:       "date only",
			targetDate: "2026-03-15",
			wantStart:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			wantEnd:    time.Date(2026, 3, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:       "RFC3339",
			targetDate: "2026-03-15T10:00:00Z",
			wantStart:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			wantEnd:    time.Date(2026, 3, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:       "invalid",
			targetDate: "not-a-date",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseDateRange(tt.targetDate)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}
