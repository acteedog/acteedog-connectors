package core

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextPatternPullRequest(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMatch  bool
		wantOwner  string
		wantRepo   string
		wantNumber string
	}{
		{
			name:       "basic pull request URL",
			input:      "https://github.com/octocat/Hello-World/pull/42",
			wantMatch:  true,
			wantOwner:  "octocat",
			wantRepo:   "Hello-World",
			wantNumber: "42",
		},
		{
			name:       "pull request URL with trailing slash",
			input:      "https://github.com/octocat/Hello-World/pull/42/",
			wantMatch:  true,
			wantOwner:  "octocat",
			wantRepo:   "Hello-World",
			wantNumber: "42",
		},
		{
			name:       "pull request URL in Slack mrkdwn format",
			input:      "<https://github.com/octocat/Hello-World/pull/42|PR #42>",
			wantMatch:  true,
			wantOwner:  "octocat",
			wantRepo:   "Hello-World",
			wantNumber: "42",
		},
		{
			name:      "no match for issue URL",
			input:     "https://github.com/octocat/Hello-World/issues/42",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(ContextPatternPullRequest)
			match := re.FindStringSubmatch(tt.input)

			if !tt.wantMatch {
				assert.Nil(t, match)
				return
			}

			params := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if i != 0 && name != "" {
					params[name] = match[i]
				}
			}

			assert.Equal(t, tt.wantOwner, params["owner"])
			assert.Equal(t, tt.wantRepo, params["repo"])
			assert.Equal(t, tt.wantNumber, params["number"])
		})
	}
}

func TestContextPatternIssue(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMatch  bool
		wantOwner  string
		wantRepo   string
		wantNumber string
	}{
		{
			name:       "basic issue URL",
			input:      "https://github.com/octocat/Hello-World/issues/42",
			wantMatch:  true,
			wantOwner:  "octocat",
			wantRepo:   "Hello-World",
			wantNumber: "42",
		},
		{
			name:       "issue URL with trailing slash",
			input:      "https://github.com/octocat/Hello-World/issues/42/",
			wantMatch:  true,
			wantOwner:  "octocat",
			wantRepo:   "Hello-World",
			wantNumber: "42",
		},
		{
			name:       "pull request URL in Slack mrkdwn format",
			input:      "<https://github.com/octocat/Hello-World/issues/42|Issue #42>",
			wantMatch:  true,
			wantOwner:  "octocat",
			wantRepo:   "Hello-World",
			wantNumber: "42",
		},
		{
			name:      "no match for pull request URL",
			input:     "https://github.com/octocat/Hello-World/pull/42",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(ContextPatternIssue)
			match := re.FindStringSubmatch(tt.input)

			if !tt.wantMatch {
				assert.Nil(t, match)
				return
			}

			params := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if i != 0 && name != "" {
					params[name] = match[i]
				}
			}

			assert.Equal(t, tt.wantOwner, params["owner"])
			assert.Equal(t, tt.wantRepo, params["repo"])
			assert.Equal(t, tt.wantNumber, params["number"])
		})
	}
}

func TestContextPatternRepository(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMatch bool
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "basic repository URL",
			input:     "https://github.com/octocat/Hello-World",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL with trailing slash",
			input:     "https://github.com/octocat/Hello-World/",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL in Slack mrkdwn format",
			input:     "<https://github.com/octocat/Hello-World|My Repository>",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL end with ')'",
			input:     "https://github.com/octocat/Hello-World)",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL end with ']'",
			input:     "https://github.com/octocat/Hello-World]",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL end with '\"'",
			input:     "https://github.com/octocat/Hello-World\"",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL end with '''",
			input:     "https://github.com/octocat/Hello-World'",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
		{
			name:      "repository URL end with query parameters",
			input:     "https://github.com/octocat/Hello-World?param=value",
			wantMatch: true,
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(ContextPatternRepository)
			match := re.FindStringSubmatch(tt.input)

			if !tt.wantMatch {
				assert.Nil(t, match)
				return
			}

			params := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if i != 0 && name != "" {
					params[name] = match[i]
				}
			}

			assert.Equal(t, tt.wantOwner, params["owner"])
			assert.Equal(t, tt.wantRepo, params["repo"])
		})
	}
}
