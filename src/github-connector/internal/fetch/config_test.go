package fetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name       string
		cfg        map[string]any
		targetDate string
		wantConfig *config
		wantErr    bool
	}{
		{
			name: "valid config",
			cfg: map[string]any{
				"credential_personal_access_token": "valid_token",
				"username":                         "octocat",
				"repository_patterns":              []any{"octocat/*"},
			},
			targetDate: "2025-12-12T12:00:00+09:00",
			wantConfig: &config{
				token:              "valid_token",
				username:           "octocat",
				repositoryPatterns: []string{"octocat/*"},
				startTime:          time.Date(2025, 12, 12, 0, 0, 0, 0, time.UTC),
				endTime:            time.Date(2025, 12, 12, 23, 59, 59, 999999999, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "valid config - date only format",
			cfg: map[string]any{
				"credential_personal_access_token": "valid_token",
				"username":                         "octocat",
				"repository_patterns":              []any{"octocat/*"},
			},
			targetDate: "2025-12-12",
			wantConfig: &config{
				token:              "valid_token",
				username:           "octocat",
				repositoryPatterns: []string{"octocat/*"},
				startTime:          time.Date(2025, 12, 12, 0, 0, 0, 0, time.UTC),
				endTime:            time.Date(2025, 12, 12, 23, 59, 59, 999999999, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing token",
			cfg: map[string]any{
				"username":            "octocat",
				"repository_patterns": []any{"octocat/*"},
			},
			targetDate: "2025-12-12T12:00:00+09:00",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing username",
			cfg: map[string]any{
				"credential_personal_access_token": "valid_token",
				"repository_patterns":              []any{"octocat/*"},
			},
			targetDate: "2025-12-12T12:00:00+09:00",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - invalid target date",
			cfg: map[string]any{
				"credential_personal_access_token": "valid_token",
				"username":                         "octocat",
				"repository_patterns":              []any{"octocat/*"},
			},
			targetDate: "invalid date format",
			wantConfig: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := newConfig(tt.cfg, tt.targetDate)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantConfig, gotConfig)
		})
	}
}
