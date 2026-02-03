package main

import "github-connector/internal/core"

// GetContextPatterns returns URL patterns and context mappings for cross-plugin detection
func GetContextPatterns() (ContextPatternsResponse, error) {
	return ContextPatternsResponse{
		Patterns: []ContextPatternDefinition{
			// Pull Request URL pattern
			{
				Pattern: `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/pull/(?P<number>\d+)`,
				ContextMappings: []ContextMapping{
					{
						NameTemplate:     "github:source",
						Level:            1,
						ResourceType:     core.ResourceTypeSource,
						ParentIndex:      nil,
						IdTemplate:       core.MakeSourceContextID(),
						EnrichmentParams: map[string]any{},
					},
					{
						NameTemplate: "repository:{{owner}}/{{repo}}",
						Level:        2,
						ResourceType: core.ResourceTypeRepository,
						ParentIndex:  int64Ptr(0),
						IdTemplate:   core.MakeRepositoryContextID("{{owner}}/{{repo}}"),
						EnrichmentParams: map[string]any{
							"repo": "{{owner}}/{{repo}}",
						},
					},
					{
						NameTemplate: "PR #{{number}}",
						Level:        3,
						ResourceType: core.ResourceTypePullRequest,
						ParentIndex:  int64Ptr(1),
						IdTemplate:   core.MakePullRequestContextID("{{owner}}/{{repo}}", "{{number}}"),
						EnrichmentParams: map[string]any{
							"repo":      "{{owner}}/{{repo}}",
							"pr_number": "{{number}}",
						},
					},
				},
			},
			// Issue URL pattern
			{
				Pattern: `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/issues/(?P<number>\d+)`,
				ContextMappings: []ContextMapping{
					{
						NameTemplate:     "github:source",
						Level:            1,
						ResourceType:     core.ResourceTypeSource,
						ParentIndex:      nil,
						IdTemplate:       core.MakeSourceContextID(),
						EnrichmentParams: map[string]any{},
					},
					{
						NameTemplate: "repository:{{owner}}/{{repo}}",
						Level:        2,
						ResourceType: core.ResourceTypeRepository,
						ParentIndex:  int64Ptr(0),
						IdTemplate:   core.MakeRepositoryContextID("{{owner}}/{{repo}}"),
						EnrichmentParams: map[string]any{
							"repo": "{{owner}}/{{repo}}",
						},
					},
					{
						NameTemplate: "Issue #{{number}}",
						Level:        3,
						ResourceType: core.ResourceTypeIssue,
						ParentIndex:  int64Ptr(1),
						IdTemplate:   core.MakeIssueContextID("{{owner}}/{{repo}}", "{{number}}"),
						EnrichmentParams: map[string]any{
							"repo":         "{{owner}}/{{repo}}",
							"issue_number": "{{number}}",
						},
					},
				},
			},
			// Repository URL pattern
			{
				Pattern: `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/?$`,
				ContextMappings: []ContextMapping{
					{
						NameTemplate:     "github:source",
						Level:            1,
						ResourceType:     core.ResourceTypeSource,
						ParentIndex:      nil,
						IdTemplate:       core.MakeSourceContextID(),
						EnrichmentParams: map[string]any{},
					},
					{
						NameTemplate: "repository:{{owner}}/{{repo}}",
						Level:        2,
						ResourceType: core.ResourceTypeRepository,
						ParentIndex:  int64Ptr(0),
						IdTemplate:   core.MakeRepositoryContextID("{{owner}}/{{repo}}"),
						EnrichmentParams: map[string]any{
							"repo": "{{owner}}/{{repo}}",
						},
					},
				},
			},
		},
	}, nil
}

// int64Ptr returns a pointer to an int64
func int64Ptr(i int64) *int64 {
	return &i
}
