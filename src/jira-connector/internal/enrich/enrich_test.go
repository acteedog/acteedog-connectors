package enrich

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
	project *JiraProjectResponse
	issue   *JiraIssueResponse
	err     error
}

func (m *mockHTTPClient) FetchProject(cloudID, email, apiToken, projectID string) (*JiraProjectResponse, error) {
	return m.project, m.err
}

func (m *mockHTTPClient) FetchIssue(cloudID, email, apiToken, issueID string) (*JiraIssueResponse, error) {
	return m.issue, m.err
}

func loadJiraProjectResponse(t *testing.T, path string) *JiraProjectResponse {
	t.Helper()

	b, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	var data JiraProjectResponse
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	return &data
}

func loadJiraIssueResponse(t *testing.T, path string) *JiraIssueResponse {
	t.Helper()

	b, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	var data JiraIssueResponse
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	return &data
}

func ptrString(s string) *string {
	return &s
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func TestEnrichContext(t *testing.T) {
	cfg := map[string]any{
		"cloud_id":       "cloud-id",
		"email":          "test.user@example.com",
		"api_token":      "test-api-token",
		"project_ids":    []any{"10000"},
		"site_subdomain": "myorg",
	}

	tests := []struct {
		name           string
		projectFixture string
		issueFixture   string
		contextType    string
		params         map[string]any
		want           *core.Context
		wantErr        bool
	}{
		{
			name:        "source",
			contextType: "source",
			params:      map[string]any{},
			want: &core.Context{
				Title:       ptrString("Jira"),
				Description: ptrString("Activity source from Jira"),
				Url:         ptrString("https://api.atlassian.com/ex/jira/cloud-id"),
			},
			wantErr: false,
		},
		{
			name:           "project",
			projectFixture: "../../testdata/enrichment/project.json",
			contextType:    "project",
			params: map[string]any{
				"project_id": "10000",
			},
			want: &core.Context{
				Title:       ptrString("test-project"),
				Description: ptrString("Test project"),
				Url:         ptrString("https://myorg.atlassian.net/jira/software/projects/TES/boards"),
				Metadata: map[string]any{
					"key":              "TES",
					"project_type_key": "software",
				},
			},
			wantErr: false,
		},
		{
			name:         "issue",
			issueFixture: "../../testdata/enrichment/issue.json",
			contextType:  "issue",
			params: map[string]any{
				"issue_id": "10038",
			},
			want: &core.Context{
				Title:       ptrString("Test Epic"),
				Description: ptrString("This is a test epic."),
				Url:         ptrString("https://myorg.atlassian.net/browse/TES-6"),
				// "2026-03-10T23:04:42.979+0900" → UTC
				CreatedAt: ptrTime(time.Date(2026, 3, 10, 14, 4, 42, 979000000, time.UTC)),
				// "2026-03-10T23:04:44.068+0900" → UTC
				UpdatedAt: ptrTime(time.Date(2026, 3, 10, 14, 4, 44, 68000000, time.UTC)),
				Metadata: map[string]any{
					"issue_type": "Epic",
					"status":     "To Do",
					"priority":   "Medium",
					"creator":    "test.user@example.com",
					"created":    "2026-03-10T23:04:42.979+0900",
					"updated":    "2026-03-10T23:04:44.068+0900",
				},
			},
			wantErr: false,
		},
		{
			name:        "unsupported type",
			contextType: "unknown",
			params:      map[string]any{},
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockHTTPClient{}
			if tt.projectFixture != "" {
				mock.project = loadJiraProjectResponse(t, tt.projectFixture)
			}
			if tt.issueFixture != "" {
				mock.issue = loadJiraIssueResponse(t, tt.issueFixture)
			}

			enricher, err := NewContextEnricher(mock, tt.contextType, cfg, tt.params, core.NewNoopLogger())
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
