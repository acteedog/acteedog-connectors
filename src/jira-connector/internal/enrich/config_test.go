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
			contextType: "issue",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			params:  map[string]any{},
			wantErr: false,
		},
		{
			name:        "valid config with enrichment params",
			contextType: "issue",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			params: map[string]any{
				"issue_id": "10038",
			},
			wantErr: false,
		},
		{
			name:        "invalid config - missing cloud_id",
			contextType: "issue",
			cfg: map[string]any{
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name:        "invalid config - missing email",
			contextType: "project",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name:        "invalid config - missing api_token",
			contextType: "project",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name:        "invalid config - missing project_ids",
			contextType: "project",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"site_subdomain": "myorg",
			},
			params:     map[string]any{},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name:        "invalid config - missing site_subdomain",
			contextType: "project",
			cfg: map[string]any{
				"cloud_id":    "my-cloud-id",
				"email":       "user@example.com",
				"api_token":   "my-api-token",
				"project_ids": []any{"10000"},
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
				assert.Nil(t, gotConfig)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotConfig)
				assert.Equal(t, tt.contextType, gotConfig.contextType)
			}
		})
	}
}
