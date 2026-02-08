package fetch

import (
	"testing"

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
				"bot_token":     "valid_token",
				"workspace_url": "example.slack.com",
				"user_id":       "U12345678",
			},
			targetDate: "2025-12-12T12:00:00+09:00",
			wantConfig: &config{
				token:        "valid_token",
				userID:       "U12345678",
				targetDate:   "2025-12-12T12:00:00+09:00",
				workspaceURL: "example.slack.com",
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing bot_token",
			cfg: map[string]any{
				"workspace_url": "example.slack.com",
				"user_id":       "U12345678",
			},
			targetDate: "2025-12-12T12:00:00+09:00",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing workspace_url",
			cfg: map[string]any{
				"bot_token": "valid_token",
				"user_id":   "U12345678",
			},
			targetDate: "2025-12-12T12:00:00+09:00",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing user_id",
			cfg: map[string]any{
				"bot_token":     "valid_token",
				"workspace_url": "example.slack.com",
			},
			targetDate: "2025-12-12T12:00:00+09:00",
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
