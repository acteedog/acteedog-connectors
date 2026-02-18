package fetch

import (
	"encoding/json"
	"os"
	"slack-connector/internal/core"
	mock_fetch "slack-connector/mock/fetch"
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

func TestFetchActivities(t *testing.T) {
	tests := []struct {
		name        string
		getMockHTTP func(*gomock.Controller) HTTPClient
		cfg         map[string]any
		targetDate  string
		want        []*Activity
		wantErr     bool
	}{
		{
			name: "thread without replies",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/thread_without_reply.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchMessages("token", "U12345678", "2025-12-13", 1).Return(map[string]any{
					"messages": map[string]any{
						"matches": []any{response},
					},
				}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"bot_token":     "token",
				"workspace_url": "test-workspace.slack.com",
				"user_id":       "U12345678",
			},
			targetDate: "2025-12-13",
			want: []*Activity{
				{
					ActivityType: "message",
					Source:       "slack",
					Id:           "slack:1765611321.248519",
					Title:        "Message in #general",
					Description:  ptrString("単体メッセージ"),
					Url:          ptrString("https://test-workspace.slack.com/archives/C099VUEKVBN/p1765611321248519"),
					Timestamp:    time.Date(2025, 12, 13, 7, 35, 21, 248518943, time.UTC),
					Metadata: map[string]any{
						"channel_id":   "C099VUEKVBN",
						"channel_name": "general",
						"user":         "sd099rsefgdb_user",
						"thread_ts":    "1765611321.248519",
						"team":         "T099VUE950C",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "slack",
							Id:           "slack:source",
							Title:        ptrString("Slack"),
							Name:         "slack:source",
							Description:  ptrString("Activity source from Slack"),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://slack.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "slack",
							Id:           "slack:channel:C099VUEKVBN",
							Title:        ptrString("#general"),
							Name:         "channel #general",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "slack:source",
							ResourceType: "channel",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"channel_id": "C099VUEKVBN",
								},
							},
						},
						{
							ConnectorId:  "slack",
							Id:           "slack:thread:C099VUEKVBN:1765611321.248519",
							Title:        ptrString("Thread 1765611321.248519"),
							Name:         "Thread 1765611321.248519",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "slack:channel:C099VUEKVBN",
							ResourceType: "thread",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"channel_id": "C099VUEKVBN",
									"thread_ts":  "1765611321.248519",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "reply",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/reply.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchMessages("token", "U12345678", "2025-12-13", 1).Return(map[string]any{
					"messages": map[string]any{
						"matches": []any{response},
					},
				}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"bot_token":     "token",
				"workspace_url": "test-workspace.slack.com",
				"user_id":       "U12345678",
			},
			targetDate: "2025-12-13",
			want: []*Activity{
				{
					ActivityType: "message",
					Source:       "slack",
					Id:           "slack:1765613227.980829",
					Title:        "Message in #general",
					Description:  ptrString("リプライです"),
					Url:          ptrString("https://test-workspace.slack.com/archives/C099VUEKVBN/p1765613227980829?thread_ts=1765613134.990399"),
					Timestamp:    time.Date(2025, 12, 13, 8, 7, 7, 980829000, time.UTC),
					Metadata: map[string]any{
						"channel_id":   "C099VUEKVBN",
						"channel_name": "general",
						"user":         "sd099rsefgdb_user",
						"thread_ts":    "1765613134.990399",
						"team":         "T099VUE950C",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "slack",
							Id:           "slack:source",
							Title:        ptrString("Slack"),
							Name:         "slack:source",
							Description:  ptrString("Activity source from Slack"),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://slack.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "slack",
							Id:           "slack:channel:C099VUEKVBN",
							Title:        ptrString("#general"),
							Name:         "channel #general",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "slack:source",
							ResourceType: "channel",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"channel_id": "C099VUEKVBN",
								},
							},
						},
						{
							ConnectorId:  "slack",
							Id:           "slack:thread:C099VUEKVBN:1765613134.990399",
							Title:        ptrString("Thread 1765613134.990399"),
							Name:         "Thread 1765613134.990399",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "slack:channel:C099VUEKVBN",
							ResourceType: "thread",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"channel_id": "C099VUEKVBN",
									"thread_ts":  "1765613134.990399",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			mockHTTP := tt.getMockHTTP(ctrl)
			fetcher, err := NewActivityFetcher(mockHTTP, tt.cfg, tt.targetDate, core.NewNoopLogger())
			if err != nil {
				t.Fatalf("Failed to create ContextEnricher: %v", err)
			}

			got, err := fetcher.FetchActivities()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func ptrString(s string) *string {
	return &s
}
