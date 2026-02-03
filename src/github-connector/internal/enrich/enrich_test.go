package enrich

import (
	"encoding/json"
	"github-connector/internal/core"
	mock_enrich "github-connector/mock/enrich"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func loadJSONTestData(t *testing.T, path string) map[string]any {
	t.Helper()

	b, err := os.ReadFile(path)
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
				"credential_personal_access_token": "token",
			},
			params: map[string]any{},
			want: &core.Context{
				Title:       ptrString("GitHub"),
				Description: ptrString("Github is a code hosting platform for version control and collaboration."),
				Url:         ptrString("https://github.com"),
			},
			wantErr: false,
		},
		{
			name: "enrich repository context",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/enrichment/repository.json")

				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchRepository("token", "owner/repo").Return(response, nil).Times(1)
				return mockHTTP
			},
			resourceType: "repository",
			cfg: map[string]any{
				"credential_personal_access_token": "token",
			},
			params: map[string]any{
				"repo": "owner/repo",
			},
			want: &core.Context{
				Title:       ptrString("Repository: testorg/testrepo"),
				Description: ptrString("My awesome test repo"),
				Url:         ptrString("https://github.com/testorg/testrepo"),
				CreatedAt:   ptrTime(time.Date(2015, 2, 13, 7, 54, 25, 0, time.UTC)),
				UpdatedAt:   ptrTime(time.Date(2025, 12, 5, 10, 30, 1, 0, time.UTC)),
				Metadata: map[string]any{
					"stargazers_count":  float64(13),
					"language":          "Go",
					"topics":            []any{"ruby-on-rails"},
					"default_branch":    "main",
					"visibility":        "public",
					"forks_count":       float64(216),
					"open_issues_count": float64(216),
					"watchers_count":    float64(13),
					"homepage":          "https://example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "enrich pull request context",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/enrichment/pr.json")

				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchPullRequest("token", "owner/repo", "123").Return(response, nil).Times(1)
				return mockHTTP
			},
			resourceType: "pull_request",
			cfg: map[string]any{
				"credential_personal_access_token": "token",
			},
			params: map[string]any{
				"repo":      "owner/repo",
				"pr_number": "123",
			},
			want: &core.Context{
				Title:       ptrString("Fix: remove all test cases"),
				Description: ptrString("This is a body of the PR."),
				Url:         ptrString("https://github.com/testorg/testrepo/pull/52742"),
				CreatedAt:   ptrTime(time.Date(2025, 11, 11, 0, 52, 36, 0, time.UTC)),
				UpdatedAt:   ptrTime(time.Date(2025, 11, 13, 5, 34, 50, 0, time.UTC)),
				Metadata: map[string]any{
					"state":         "closed",
					"author":        "john",
					"assignees":     []string{"john"},
					"reviewers":     []string{"reviewer1", "reviewer2"},
					"labels":        []string{"label1", "label2"},
					"base_branch":   "main",
					"head_branch":   "feature/awesome-branch",
					"milestone":     "",
					"additions":     float64(1),
					"deletions":     float64(998),
					"changed_files": float64(29),
					"commits_count": float64(1),
					"merged":        true,
					"merged_at":     "2025-11-13T05:34:49Z",
					"merged_by":     "john",
				},
			},
			wantErr: false,
		},
		{
			name: "enrich issue context",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/enrichment/issue.json")

				mockHTTP := mock_enrich.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchIssue("token", "owner/repo", "123").Return(response, nil).Times(1)
				return mockHTTP
			},
			resourceType: "issue",
			cfg: map[string]any{
				"credential_personal_access_token": "token",
			},
			params: map[string]any{
				"repo":         "owner/repo",
				"issue_number": "123",
			},
			want: &core.Context{
				Title:       ptrString("Use `j` and `k` for navigation in trace timeline"),
				Description: ptrString("It would be great if you could use `j` and `k` on the \"Trace Timeline\" page as well. Currently, it works on the \"main\" trace page, but not in the timeline."),
				Url:         ptrString("https://github.com/ymtdzzz/otel-tui/issues/340"),
				CreatedAt:   ptrTime(time.Date(2025, 10, 6, 19, 55, 39, 0, time.UTC)),
				UpdatedAt:   ptrTime(time.Date(2025, 11, 8, 7, 32, 14, 0, time.UTC)),
				Metadata: map[string]any{
					"state":     "open",
					"author":    "testuser",
					"assignees": []string{},
					"labels":    []string{"bug", "enhancement", "good first issue", "UI", "signal: traces"},
					"milestone": "",
					"comments":  float64(2),
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
				"credential_personal_access_token": "token",
			},
			params:  map[string]any{},
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
