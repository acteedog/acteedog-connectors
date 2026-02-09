package enrich

import (
	"encoding/json"
	"os"
	"slack-connector/internal/core"
	mock_enrich "slack-connector/mock/enrich"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func loadJSONTestData(t *testing.T, path string) map[string]any {
	t.Helper()

	b, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	var data map[string]any
	err = json.Unmarshal(b, &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	return data
}

func TestEnrichContext(t *testing.T) {
	tests := []struct {
		name         string
		getMockHTTP  func(*gomock.Controller) HTTPClient
		resourceType string
		cfg, params  map[string]any
		want         *core.Context
		wantErr      bool
	}{
		{
			name: "enrich source context",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				return mockHTTP
			},
			resourceType: "source",
			cfg: map[string]any{
				"bot_token":     "token",
				"workspace_url": "example.slack.com",
			},
			params: map[string]any{},
			want: &core.Context{
				Title:       ptrString("Slack"),
				Description: ptrString("Activity source from Slack"),
				Url:         ptrString("https://example.slack.com"),
			},
			wantErr: false,
		},
		{
			name: "enrich channel context",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/enrichment/channel.json")

				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchChannel("token", "C099VUEKVBN").Return(response, nil).Times(1)
				return mockHTTP
			},
			resourceType: "channel",
			cfg: map[string]any{
				"bot_token":     "token",
				"workspace_url": "example.slack.com",
			},
			params: map[string]any{
				"channel_id": "C099VUEKVBN",
			},
			want: &core.Context{
				Title:       ptrString("#general"),
				Description: ptrString("Company-wide announcements and work-based matters"),
				Url:         ptrString("https://example.slack.com/archives/C099VUEKVBN"),
				CreatedAt:   ptrTime(time.Date(2025, 8, 11, 2, 2, 56, 0, time.UTC)),
				UpdatedAt:   ptrTime(time.Date(2025, 12, 8, 14, 33, 49, 239000000, time.UTC)),
				Metadata: map[string]any{
					"name":            "general",
					"is_private":      false,
					"is_channel":      true,
					"is_group":        false,
					"is_im":           false,
					"topic":           "Company-wide announcements and work-based matters",
					"purpose":         "This channel is for workspace-wide communication and announcements. All members are in this channel.",
					"context_team_id": "T099VUE950C",
				},
			},
			wantErr: false,
		},
		{
			name: "enrich thread context",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/enrichment/thread.json")

				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchThread("token", "C099VUEKVBN", "1765613134.990399").Return(response, nil).Times(1)
				return mockHTTP
			},
			resourceType: "thread",
			cfg: map[string]any{
				"bot_token":     "token",
				"workspace_url": "example.slack.com",
			},
			params: map[string]any{
				"channel_id": "C099VUEKVBN",
				"thread_ts":  "1765613134.990399",
			},
			want: &core.Context{
				Title:       ptrString("Thread: 後からぶら下げる"),
				Description: ptrString("後からぶら下げる"),
				Url:         ptrString("https://example.slack.com/archives/C099VUEKVBN/p1765613134990399"),
				CreatedAt:   ptrTime(time.Date(2025, 12, 13, 8, 5, 34, 990399000, time.UTC)),
				UpdatedAt:   ptrTime(time.Date(2025, 12, 13, 8, 5, 34, 990399000, time.UTC)),
				Metadata: map[string]any{
					"parent_user":       "U099SQHSJCW",
					"parent_ts":         "1765613134.990399",
					"thread_ts":         "1765613134.990399",
					"team":              "T099VUE950C",
					"reply_count":       int64(1),
					"reply_users_count": int64(1),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid resource type",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				return mockHTTP
			},
			resourceType: "invalid resource",
			cfg: map[string]any{
				"bot_token":     "token",
				"workspace_url": "example.slack.com",
			},
			params: map[string]any{
				"channel_id": "C099VUEKVBN",
				"thread_ts":  "1765613134.990399",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			mockHTTP := tt.getMockHTTP(ctrl)
			enricher, err := NewContextEnricher(mockHTTP, tt.resourceType, tt.cfg, tt.params, core.NewNoopLogger())
			if err != nil {
				t.Fatalf("Failed to create ContextEnricher: %v", err)
			}

			got, err := enricher.EnrichContext(&core.Context{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
