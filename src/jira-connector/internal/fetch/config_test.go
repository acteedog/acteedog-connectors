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
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10T12:00:00+09:00",
			wantConfig: &config{
				startTime:  time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
				endTime:    time.Date(2026, 3, 10, 23, 59, 59, 999999999, time.UTC),
				targetDate: "2026-03-10",
			},
			wantErr: false,
		},
		{
			name: "valid config - date only format",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			wantConfig: &config{
				startTime:  time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
				endTime:    time.Date(2026, 3, 10, 23, 59, 59, 999999999, time.UTC),
				targetDate: "2026-03-10",
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing cloud_id",
			cfg: map[string]any{
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing email",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing api_token",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing project_ids",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - missing site_subdomain",
			cfg: map[string]any{
				"cloud_id":    "my-cloud-id",
				"email":       "user@example.com",
				"api_token":   "my-api-token",
				"project_ids": []any{"10000"},
			},
			targetDate: "2026-03-10",
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config - invalid target date",
			cfg: map[string]any{
				"cloud_id":       "my-cloud-id",
				"email":          "user@example.com",
				"api_token":      "my-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
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
				assert.Nil(t, gotConfig)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotConfig)
				assert.Equal(t, tt.wantConfig.startTime, gotConfig.startTime)
				assert.Equal(t, tt.wantConfig.endTime, gotConfig.endTime)
				assert.Equal(t, tt.wantConfig.targetDate, gotConfig.targetDate)
			}
		})
	}
}
