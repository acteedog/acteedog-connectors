package fetch

import (
	"encoding/json"
	"github-connector/internal/core"
	mock_fetch "github-connector/mock/fetch"
	"os"
	"testing"

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
			name: "delete event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/delete.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-18",
			want: []*Activity{
				{
					ActivityType: "delete",
					Source:       "github",
					Id:           "github:6031147775",
					Title:        "Deleted branch chore/golangci-lint in ymtdzzz/otel-tui",
					Description:  ptrString("branch chore/golangci-lint was deleted"),
					Url:          ptrString("https://github.com/ymtdzzz/otel-tui"),
					Timestamp:    "2025-11-18T00:37:03Z",
					Metadata: map[string]any{
						"ref_type":    "branch",
						"ref":         "chore/golangci-lint",
						"deleted_by":  "ymtdzzz",
						"pusher_type": "user",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:ymtdzzz/otel-tui",
							Title:        ptrString("ymtdzzz/otel-tui"),
							Name:         "repository:ymtdzzz/otel-tui",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "ymtdzzz/otel-tui",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "issue_comment event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/issue_comment.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-08",
			want: []*Activity{
				{
					ActivityType: "issue_comment",
					Source:       "github",
					Id:           "github:4548540416",
					Title:        "Commented on Issue #340",
					Description:  ptrString("I’m currently working on #352 . Once that’s done, I plan to work on this issue next."),
					Url:          ptrString("https://github.com/ymtdzzz/otel-tui/issues/340#issuecomment-3506104349"),
					Timestamp:    "2025-11-08T07:32:14Z",
					Metadata: map[string]any{
						"comment_id":         float64(3506104349),
						"issue_number":       340,
						"comment_author":     "ymtdzzz",
						"comment_created_at": "2025-11-08T07:32:14Z",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:ymtdzzz/otel-tui",
							Title:        ptrString("ymtdzzz/otel-tui"),
							Name:         "repository:ymtdzzz/otel-tui",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "ymtdzzz/otel-tui",
								},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:issue:ymtdzzz/otel-tui:340",
							Title:        ptrString("Issue #340"),
							Name:         "Issue #340",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "github:repository:ymtdzzz/otel-tui",
							ResourceType: "issue",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo":         "ymtdzzz/otel-tui",
									"issue_number": "340",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "issue event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/issues.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-17",
			want: []*Activity{
				{
					ActivityType: "issues",
					Source:       "github",
					Id:           "github:4716363699",
					Title:        "Issue #375 labeled in ymtdzzz/otel-tui",
					Description:  ptrString("Issue #375 was labeled"),
					Url:          ptrString("https://github.com/ymtdzzz/otel-tui/issues/375"),
					Timestamp:    "2025-11-17T07:15:21Z",
					Metadata: map[string]any{
						"issue_number": 375,
						"action":       "labeled",
						"state":        "open",
						"author":       "ymtdzzz",
						"labels":       []string{"CI/CD"},
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:ymtdzzz/otel-tui",
							Title:        ptrString("ymtdzzz/otel-tui"),
							Name:         "repository:ymtdzzz/otel-tui",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "ymtdzzz/otel-tui",
								},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:issue:ymtdzzz/otel-tui:375",
							Title:        ptrString("Issue #375"),
							Name:         "Issue #375",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "github:repository:ymtdzzz/otel-tui",
							ResourceType: "issue",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo":         "ymtdzzz/otel-tui",
									"issue_number": "375",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pr_comment event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/pr_comment.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-17",
			want: []*Activity{
				{
					ActivityType: "pr_comment",
					Source:       "github",
					Id:           "github:4719356430",
					Title:        "Commented on PR #52580",
					Description:  ptrString("This is a comment on the PR."),
					Url:          ptrString("https://github.com/testorg/testrepo/pull/52580#issuecomment-3540831349"),
					Timestamp:    "2025-11-17T09:43:25Z",
					Metadata: map[string]any{
						"comment_id":         float64(3540831349),
						"issue_number":       52580,
						"comment_author":     "ymtdzzz",
						"comment_created_at": "2025-11-17T09:43:25Z",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:testorg/testrepo",
							Title:        ptrString("testorg/testrepo"),
							Name:         "repository:testorg/testrepo",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "testorg/testrepo",
								},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:pull_request:testorg/testrepo:52580",
							Title:        ptrString("PR #52580"),
							Name:         "PR #52580",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "github:repository:testorg/testrepo",
							ResourceType: "pull_request",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo":      "testorg/testrepo",
									"pr_number": "52580",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pr_review event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/pr_review.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-13",
			want: []*Activity{
				{
					ActivityType: "pr_review",
					Source:       "github",
					Id:           "github:4645219660",
					Title:        "Reviewed PR #52742 in testorg/testrepo",
					Description:  ptrString("This is a great change! Approved."),
					Url:          ptrString("https://github.com/testorg/testrepo/pull/52742#pullrequestreview-3457546608"),
					Timestamp:    "2025-11-13T05:34:50Z",
					Metadata: map[string]any{
						"pr_number":    52742,
						"review_state": "approved",
						"reviewer":     "ymtdzzz",
						"submitted_at": "2025-11-13T05:34:48Z",
						"base_branch":  "main",
						"head_branch":  "feature/my-awesome-feature",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:testorg/testrepo",
							Title:        ptrString("testorg/testrepo"),
							Name:         "repository:testorg/testrepo",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "testorg/testrepo",
								},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:pull_request:testorg/testrepo:52742",
							Title:        ptrString("PR #52742"),
							Name:         "PR #52742",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "github:repository:testorg/testrepo",
							ResourceType: "pull_request",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo":      "testorg/testrepo",
									"pr_number": "52742",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pr_review_comment event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/pr_review_comment.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-17",
			want: []*Activity{
				{
					ActivityType: "pr_review_comment",
					Source:       "github",
					Id:           "github:4714496145",
					Title:        "Commented on PR #52580 in testorg/testrepo",
					Description:  ptrString("This is pull request review comment body."),
					Url:          ptrString("https://github.com/testorg/testrepo/pull/52580#discussion_r2532700510"),
					Timestamp:    "2025-11-17T04:56:32Z",
					Metadata: map[string]any{
						"comment_id":     float64(2532700510),
						"pr_number":      52580,
						"comment_author": "ymtdzzz",
						"file_path":      "path/to/file.rb",
						"commit_id":      "b3e292a848f0547955e4f2d9b4803acb70faefd1",
						"base_branch":    "main",
						"head_branch":    "feature/awesome-feature",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:testorg/testrepo",
							Title:        ptrString("testorg/testrepo"),
							Name:         "repository:testorg/testrepo",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "testorg/testrepo",
								},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:pull_request:testorg/testrepo:52580",
							Title:        ptrString("PR #52580"),
							Name:         "PR #52580",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "github:repository:testorg/testrepo",
							ResourceType: "pull_request",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo":      "testorg/testrepo",
									"pr_number": "52580",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pull_request event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/pull_request.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-12",
			want: []*Activity{
				{
					ActivityType: "pull_request",
					Source:       "github",
					Id:           "github:1234567890",
					Title:        "PR #10286 merged in testorg/testrepo",
					Description:  ptrString("Pull request #10286 was merged"),
					Url:          ptrString("https://github.com/testorg/testrepo/pull/10286"),
					Timestamp:    "2025-11-12T01:06:01Z",
					Metadata: map[string]any{
						"pr_number":   10286,
						"action":      "merged",
						"base_branch": "main",
						"head_branch": "feature/my-awesome-feature",
						"base_sha":    "4d0ac009a8e1f363fb6fea838abc52b2351d184e",
						"head_sha":    "556eadf823c287022e62c8d76b77fe24371080f6",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:testorg/testrepo",
							Title:        ptrString("testorg/testrepo"),
							Name:         "repository:testorg/testrepo",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "testorg/testrepo",
								},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:pull_request:testorg/testrepo:10286",
							Title:        ptrString("PR #10286"),
							Name:         "PR #10286",
							Description:  nil,
							Level:        int64(3),
							ParentId:     "github:repository:testorg/testrepo",
							ResourceType: "pull_request",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo":      "testorg/testrepo",
									"pr_number": "10286",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "push event",
			getMockHTTP: func(ctrl *gomock.Controller) HTTPClient {
				response := loadJSONTestData(t, "../../testdata/events/push.json")

				mockHTTP := mock_fetch.NewMockHTTPClient(ctrl)
				mockHTTP.EXPECT().FetchActivities("token", "username", 1).Return([]map[string]any{response}, nil).Times(1)
				mockHTTP.EXPECT().FetchActivities("token", "username", 2).Return([]map[string]any{}, nil).Times(1)
				return mockHTTP
			},
			cfg: map[string]any{
				"credential_personal_access_token": "token",
				"username":                         "username",
			},
			targetDate: "2025-11-12",
			want: []*Activity{
				{
					ActivityType: "push",
					Source:       "github",
					Id:           "github:5894071350",
					Title:        "Push to ymtdzzz/otel-tui",
					Description:  ptrString("Pushed to refs/heads/feature/refactor_components in ymtdzzz/otel-tui"),
					Url:          ptrString("https://github.com/ymtdzzz/otel-tui/commit/4fb5eb96ecc5141ff2383d720508bd0ccaa1b820"),
					Timestamp:    "2025-11-12T12:04:19Z",
					Metadata: map[string]any{
						"branch":        "refs/heads/feature/refactor_components",
						"before_commit": "61f45b540397ef414133da92c420442b5acac554",
					},
					Contexts: []*core.Context{
						{
							ConnectorId:  "github",
							Id:           "github:source",
							Title:        ptrString("GitHub"),
							Name:         "github:source",
							Description:  ptrString("Github is a code hosting platform for version control and collaboration."),
							Level:        int64(1),
							ParentId:     "",
							ResourceType: "source",
							Url:          ptrString("https://github.com"),
							Metadata: map[string]any{
								"enrichment_params": map[string]any{},
							},
						},
						{
							ConnectorId:  "github",
							Id:           "github:repository:ymtdzzz/otel-tui",
							Title:        ptrString("ymtdzzz/otel-tui"),
							Name:         "repository:ymtdzzz/otel-tui",
							Description:  nil,
							Level:        int64(2),
							ParentId:     "github:source",
							ResourceType: "repository",
							Url:          nil,
							Metadata: map[string]any{
								"enrichment_params": map[string]any{
									"repo": "ymtdzzz/otel-tui",
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
