package enrich

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		contextType string
		params      any
		wantCalID   string
		wantEvtID   string
		wantErr     bool
	}{
		{
			name:        "calendar params",
			contextType: "calendar",
			params: map[string]any{
				"calendar_id": "primary",
			},
			wantCalID: "primary",
		},
		{
			name:        "event params",
			contextType: "event",
			params: map[string]any{
				"calendar_id": "primary",
				"event_id":    "evt1",
			},
			wantCalID: "primary",
			wantEvtID: "evt1",
		},
		{
			name:        "source params (empty)",
			contextType: "source",
			params:      map[string]any{},
			wantCalID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newConfig(tt.contextType, tt.params)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.contextType, cfg.contextType)
			assert.Equal(t, tt.wantCalID, cfg.enrichmentParams.CalendarID)
			assert.Equal(t, tt.wantEvtID, cfg.enrichmentParams.EventID)
		})
	}
}
