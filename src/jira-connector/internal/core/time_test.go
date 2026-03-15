package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseJiraTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "RFC3339Nano with colon offset",
			input:   "2026-03-10T14:04:42.979+09:00",
			want:    time.Date(2026, 3, 10, 5, 4, 42, 979000000, time.UTC),
			wantErr: false,
		},
		{
			name:    "RFC3339 with colon offset",
			input:   "2026-03-10T14:04:42+09:00",
			want:    time.Date(2026, 3, 10, 5, 4, 42, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Jira format without colon in offset (nanoseconds)",
			input:   "2026-03-10T14:04:42.979+0900",
			want:    time.Date(2026, 3, 10, 5, 4, 42, 979000000, time.UTC),
			wantErr: false,
		},
		{
			name:    "Jira format without colon in offset (no nanoseconds)",
			input:   "2026-03-10T14:04:42+0900",
			want:    time.Date(2026, 3, 10, 5, 4, 42, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "UTC offset",
			input:   "2026-03-10T05:04:42.000+00:00",
			want:    time.Date(2026, 3, 10, 5, 4, 42, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJiraTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Equal(got.UTC()), "expected %v, got %v", tt.want, got.UTC())
			}
		})
	}
}
