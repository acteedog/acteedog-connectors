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
			contextType: "pull_request",
			cfg: map[string]any{
				"credential_personal_access_token": "valid_token",
			},
			params: map[string]any{},
			wantConfig: &config{
				contextType:      "pull_request",
				token:            "valid_token",
				enrichmentParams: map[string]any{},
			},
			wantErr: false,
		},
		{
			name:        "invalid config - missing token",
			contextType: "pull_request",
			cfg:         map[string]any{},
			params:      map[string]any{},
			wantConfig:  nil,
			wantErr:     true,
		},
		{
			name:        "invalid config - wrong token type",
			contextType: "pull_request",
			cfg: map[string]any{
				"credential_personal_access_token": 123,
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
