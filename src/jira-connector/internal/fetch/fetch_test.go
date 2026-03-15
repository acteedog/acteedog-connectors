package fetch

import (
	"encoding/json"
	"jira-connector/internal/core"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient is a simple in-test implementation of HTTPClient
type mockHTTPClient struct {
	response *JiraSearchResponse
	err      error
}

func (m *mockHTTPClient) FetchIssues(cloudID, email, apiToken string, projectIDs []string, dateFrom, dateTo string) (*JiraSearchResponse, error) {
	return m.response, m.err
}

func loadJiraSearchResponse(t *testing.T, path string) *JiraSearchResponse {
	t.Helper()

	b, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	var data JiraSearchResponse
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	return &data
}

func TestFetchActivities(t *testing.T) {
	tests := []struct {
		name         string
		testdataPath string
		cfg          map[string]any
		targetDate   string
		want         []*Activity
		wantErr      bool
	}{
		{
			name:         "issue_creation_without_parent",
			testdataPath: "../../testdata/events/issue_creation_without_parent.json",
			cfg: map[string]any{
				"cloud_id":       "cloud-id",
				"email":          "test.user@example.com",
				"api_token":      "test-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			want: []*Activity{
				{
					Id:           "jira:issue:10038:created",
					Source:       "jira",
					ActivityType: "created",
					Title:        "Created issue TES-6: Test Epic",
					Description:  "",
					Url:          ptrString("https://myorg.atlassian.net/browse/TES-6"),
					// "2026-03-10T23:04:42.979+0900" → UTC
					Timestamp: time.Date(2026, 3, 10, 14, 4, 42, 979000000, time.UTC),
					Metadata: map[string]any{
						"issue_id":  "10038",
						"issue_key": "TES-6",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "jira",
							Id:           "jira:source",
							Name:         "jira:source",
							ParentId:     "",
							ResourceType: "source",
							Title:        ptrString("Jira"),
							Description:  ptrString("Activity source from Jira"),
							Url:          ptrString("https://api.atlassian.com/ex/jira/cloud-id"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:project:10000",
							Name:         "project test-project",
							ParentId:     "jira:source",
							ResourceType: "project",
							Title:        ptrString("test-project"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"project_id": "10000",
								},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:issue:10038",
							Name:         "epic TES-6",
							ParentId:     "jira:project:10000",
							ResourceType: "issue",
							Title:        ptrString("Test Epic"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"issue_id": "10038",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "comment_with_parent",
			testdataPath: "../../testdata/events/comment_with_parent.json",
			cfg: map[string]any{
				"cloud_id":       "cloud-id",
				"email":          "test.user@example.com",
				"api_token":      "test-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			want: []*Activity{
				{
					// comment id=10000: "2026-03-10T22:26:30.554+0900" → UTC 13:26:30.554
					Id:           "jira:issue:10003:comment:10000",
					Source:       "jira",
					ActivityType: "commented",
					Title:        "Commented on TES-4: Test Task",
					Description:  "This is a test comment.",
					Url:          ptrString("https://myorg.atlassian.net/browse/TES-4?focusedCommentId=10000"),
					Timestamp:    time.Date(2026, 3, 10, 13, 26, 30, 554000000, time.UTC),
					Metadata: map[string]any{
						"issue_id":   "10003",
						"issue_key":  "TES-4",
						"comment_id": "10000",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "jira",
							Id:           "jira:source",
							Name:         "jira:source",
							ParentId:     "",
							ResourceType: "source",
							Title:        ptrString("Jira"),
							Description:  ptrString("Activity source from Jira"),
							Url:          ptrString("https://api.atlassian.com/ex/jira/cloud-id"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:project:10000",
							Name:         "project test-project",
							ParentId:     "jira:source",
							ResourceType: "project",
							Title:        ptrString("test-project"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"project_id": "10000",
								},
							},
						},
						{
							// parent issue TES-6
							ConnectorId:  "jira",
							Id:           "jira:issue:10038",
							Name:         "Epic TES-6",
							ParentId:     "jira:project:10000",
							ResourceType: "issue",
							Title:        ptrString("Test Epic"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"issue_id": "10038",
								},
							},
						},
						{
							// issue TES-4
							ConnectorId:  "jira",
							Id:           "jira:issue:10003",
							Name:         "Task TES-4",
							ParentId:     "jira:issue:10038",
							ResourceType: "issue",
							Title:        ptrString("Test Task"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"issue_id": "10003",
								},
							},
						},
					},
				},
				{
					// comment id=10001: "2026-03-10T23:12:48.274+0900" → UTC 14:12:48.274
					Id:           "jira:issue:10003:comment:10001",
					Source:       "jira",
					ActivityType: "commented",
					Title:        "Commented on TES-4: Test Task",
					Description:  "This is another comment.",
					Url:          ptrString("https://myorg.atlassian.net/browse/TES-4?focusedCommentId=10001"),
					Timestamp:    time.Date(2026, 3, 10, 14, 12, 48, 274000000, time.UTC),
					Metadata: map[string]any{
						"issue_id":   "10003",
						"issue_key":  "TES-4",
						"comment_id": "10001",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "jira",
							Id:           "jira:source",
							Name:         "jira:source",
							ParentId:     "",
							ResourceType: "source",
							Title:        ptrString("Jira"),
							Description:  ptrString("Activity source from Jira"),
							Url:          ptrString("https://api.atlassian.com/ex/jira/cloud-id"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:project:10000",
							Name:         "project test-project",
							ParentId:     "jira:source",
							ResourceType: "project",
							Title:        ptrString("test-project"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"project_id": "10000",
								},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:issue:10038",
							Name:         "Epic TES-6",
							ParentId:     "jira:project:10000",
							ResourceType: "issue",
							Title:        ptrString("Test Epic"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"issue_id": "10038",
								},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:issue:10003",
							Name:         "Task TES-4",
							ParentId:     "jira:issue:10038",
							ResourceType: "issue",
							Title:        ptrString("Test Task"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"issue_id": "10003",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "status_change_without_parent",
			testdataPath: "../../testdata/events/status_change_without_parent.json",
			cfg: map[string]any{
				"cloud_id":       "cloud-id",
				"email":          "test.user@example.com",
				"api_token":      "test-api-token",
				"project_ids":    []any{"10000"},
				"site_subdomain": "myorg",
			},
			targetDate: "2026-03-10",
			want: []*Activity{
				{
					// history id=10045: "2026-03-10T22:51:01.216+0900" → UTC 13:51:01.216
					Id:           "jira:issue:10003:status_changed:10045",
					Source:       "jira",
					ActivityType: "status_changed",
					Title:        "Changed status of TES-4 from To Do to Doing",
					Description:  "",
					Url:          ptrString("https://myorg.atlassian.net/browse/TES-4"),
					Timestamp:    time.Date(2026, 3, 10, 13, 51, 1, 216000000, time.UTC),
					Metadata: map[string]any{
						"issue_id":    "10003",
						"issue_key":   "TES-4",
						"from_status": "To Do",
						"to_status":   "Doing",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "jira",
							Id:           "jira:source",
							Name:         "jira:source",
							ParentId:     "",
							ResourceType: "source",
							Title:        ptrString("Jira"),
							Description:  ptrString("Activity source from Jira"),
							Url:          ptrString("https://api.atlassian.com/ex/jira/cloud-id"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:project:10000",
							Name:         "project test-project",
							ParentId:     "jira:source",
							ResourceType: "project",
							Title:        ptrString("test-project"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"project_id": "10000",
								},
							},
						},
						{
							ConnectorId:  "jira",
							Id:           "jira:issue:10003",
							Name:         "Task TES-4",
							ParentId:     "jira:project:10000",
							ResourceType: "issue",
							Title:        ptrString("Test Task"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"issue_id": "10003",
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
			response := loadJiraSearchResponse(t, tt.testdataPath)
			httpClient := &mockHTTPClient{response: response}

			fetcher, err := NewActivityFetcher(httpClient, tt.cfg, tt.targetDate, core.NewNoopLogger())
			if err != nil {
				t.Fatalf("Failed to create ActivityFetcher: %v", err)
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
