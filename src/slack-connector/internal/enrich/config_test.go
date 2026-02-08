package enrich

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		contextType string
		cfg         map[string]any
		params      map[string]any
		wantConfig  *config
		wantErr     bool
	}{
		{
			name:        "valid config",
			contextType: "channel",
			cfg: map[string]any{
				"bot_token":     "valid_token",
				"workspace_url": "https://example.slack.com",
			},
			params: map[string]any{},
			wantConfig: &config{
				contextType:      "channel",
				token:            "valid_token",
				workspaceURL:     "https://example.slack.com",
				enrichmentParams: map[string]any{},
			},
			wantErr: false,
		},
		{
			name:        "invalid config - missing token",
			contextType: "channel",
			cfg: map[string]any{
				"workspace_url": "https://example.slack.com",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name:        "invalid config - missing workspace_url",
			contextType: "channel",
			cfg: map[string]any{
				"bot_token": "valid_token",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name:        "invalid config - wrong token type",
			contextType: "channel",
			cfg: map[string]any{
				"bot_token":     123,
				"workspace_url": "https://example.slack.com",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := newConfig(tt.contextType, tt.cfg, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantConfig, gotConfig)
		})
	}
}
