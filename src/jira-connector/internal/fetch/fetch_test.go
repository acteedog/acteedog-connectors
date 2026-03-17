package fetch

import (
	"encoding/json"
	"jira-connector/internal/core"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient is a simple in-test implementation of HTTPClient.
// If responses has multiple entries, each successive call returns the next response.
// If all responses are exhausted, the last one is returned on subsequent calls.
type mockHTTPClient struct {
	responses      []*JiraSearchResponse
	errs           []error
	callCount      int
	capturedTokens []string
}

func newMockHTTPClient(response *JiraSearchResponse, err error) *mockHTTPClient {
	return &mockHTTPClient{responses: []*JiraSearchResponse{response}, errs: []error{err}}
}

func newPaginatedMockHTTPClient(responses []*JiraSearchResponse) *mockHTTPClient {
	errs := make([]error, len(responses))
	return &mockHTTPClient{responses: responses, errs: errs}
}

func (m *mockHTTPClient) FetchIssues(cloudID, email, apiToken string, projectIDs []string, dateFrom, dateTo string, nextPageToken string) (*JiraSearchResponse, error) {
	m.capturedTokens = append(m.capturedTokens, nextPageToken)
	idx := m.callCount
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.callCount++
	return m.responses[idx], m.errs[idx]
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
					Id:           "jira:project:10000:issue:10038:created",
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
							Id:           "jira:project:10000:issue:10038",
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
					Id:           "jira:project:10000:issue:10003:comment:10000",
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
							Id:           "jira:project:10000:issue:10038",
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
							Id:           "jira:project:10000:issue:10003",
							Name:         "Task TES-4",
							ParentId:     "jira:project:10000:issue:10038",
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
					Id:           "jira:project:10000:issue:10003:comment:10001",
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
							Id:           "jira:project:10000:issue:10038",
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
							Id:           "jira:project:10000:issue:10003",
							Name:         "Task TES-4",
							ParentId:     "jira:project:10000:issue:10038",
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
					Id:           "jira:project:10000:issue:10003:status_changed:10045",
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
							Id:           "jira:project:10000:issue:10003",
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
			httpClient := newMockHTTPClient(response, nil)

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

func TestFetchActivities_Pagination(t *testing.T) {
	cfg := map[string]any{
		"cloud_id":       "cloud-id",
		"email":          "test.user@example.com",
		"api_token":      "test-api-token",
		"project_ids":    []any{"10000"},
		"site_subdomain": "myorg",
	}

	// Page 1: isLast=false, nextPageToken="page2token"
	page1 := &JiraSearchResponse{
		IsLast:        false,
		NextPageToken: "page2token",
		Issues: []JiraIssue{
			{
				ID:  "10001",
				Key: "TES-1",
				Fields: JiraFields{
					Summary:   "Issue Page 1",
					Created:   "2026-03-10T10:00:00.000+0000",
					Creator:   &JiraUser{EmailAddress: "test.user@example.com"},
					Project:   &JiraProjectRef{ID: "10000", Key: "TES", Name: "test-project"},
					IssueType: &JiraIssueType{ID: "10001", Name: "Task"},
				},
			},
		},
	}

	// Page 2: isLast=true
	page2 := &JiraSearchResponse{
		IsLast: true,
		Issues: []JiraIssue{
			{
				ID:  "10002",
				Key: "TES-2",
				Fields: JiraFields{
					Summary:   "Issue Page 2",
					Created:   "2026-03-10T11:00:00.000+0000",
					Creator:   &JiraUser{EmailAddress: "test.user@example.com"},
					Project:   &JiraProjectRef{ID: "10000", Key: "TES", Name: "test-project"},
					IssueType: &JiraIssueType{ID: "10001", Name: "Task"},
				},
			},
		},
	}

	httpClient := newPaginatedMockHTTPClient([]*JiraSearchResponse{page1, page2})

	fetcher, err := NewActivityFetcher(httpClient, cfg, "2026-03-10", core.NewNoopLogger())
	if err != nil {
		t.Fatalf("Failed to create ActivityFetcher: %v", err)
	}

	got, err := fetcher.FetchActivities()
	assert.NoError(t, err)

	// 2 pages × 1 issue each → 2 created activities
	assert.Len(t, got, 2)
	assert.Equal(t, 2, httpClient.callCount)

	// 1st call should have empty token, 2nd call should have "page2token"
	assert.Equal(t, []string{"", "page2token"}, httpClient.capturedTokens)

	// Check activity IDs
	ids := []string{got[0].Id, got[1].Id}
	assert.Contains(t, ids, "jira:project:10000:issue:10001:created")
	assert.Contains(t, ids, "jira:project:10000:issue:10002:created")
}
