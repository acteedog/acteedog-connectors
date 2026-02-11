// Note: run `go doc -all` in this package to see all of the types and functions available.
// ./pdk.gen.go contains the domain types from the host where your plugin will run.
package main

import (
	"encoding/json"
	"fmt"
	"github-connector/internal/core"

	"github.com/extism/go-pdk"
)

// GetConfigSchema returns the configuration schema for the GitHub connector
func GetConfigSchema() (ConfigSchema, error) {
	return ConfigSchema{
		Type: "object",
		Properties: map[string]any{
			"credential_personal_access_token": map[string]any{
				"type":        "string",
				"title":       "Personal Access Token",
				"description": "GitHub Personal Access Token for authentication",
			},
			"username": map[string]any{
				"type":        "string",
				"title":       "Username",
				"description": "GitHub username to fetch activities for",
			},
			"repository_patterns": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"title":       "Repository Patterns",
				"description": "Repository patterns to include (e.g., 'myorg/*', 'user/repo'). Leave empty for all repositories. Use * for wildcards.",
			},
		},
		Required: &[]string{
			"credential_personal_access_token",
			"username",
		},
	}, nil
}

// TestConnection tests the GitHub API connection using the provided configuration
func TestConnection(input TestConnectionRequest) error {
	pdk.Log(pdk.LogInfo, "TestConnection: Starting GitHub API connection test")

	err := validateConfig(input.Config)
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Configuration validation failed: %v", err))
		return err
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid configuration format")
	}

	token, ok := config["credential_personal_access_token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("personal access token is required")
	}

	url := fmt.Sprintf("%s/user", core.GithubAPIBaseURL)

	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", fmt.Sprintf("token %s", token))
	req.SetHeader("Accept", "application/vnd.github.v3+json")
	req.SetHeader("User-Agent", "acteedog-github-connector")

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Testing connection to: %s", url))

	res := req.Send()
	statusCode := res.Status()

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Response status: %d", statusCode))

	if statusCode == 200 {
		pdk.Log(pdk.LogInfo, "Connection test successful")
		return nil
	}

	var errorMsg string
	if statusCode == 401 {
		errorMsg = "Authentication failed: Invalid or expired Personal Access Token"
	} else if statusCode == 403 {
		errorMsg = "Access forbidden: Check token permissions"
	} else if statusCode >= 500 {
		errorMsg = fmt.Sprintf("GitHub API server error (status %d)", statusCode)
	} else {
		errorMsg = fmt.Sprintf("Connection failed with status %d", statusCode)
	}

	var githubError struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(res.Body(), &githubError); err == nil && githubError.Message != "" {
		errorMsg = fmt.Sprintf("%s: %s", errorMsg, githubError.Message)
	}

	pdk.Log(pdk.LogError, errorMsg)

	return fmt.Errorf("%s", errorMsg)
}

// validateConfig checks required fields and repository pattern formats
func validateConfig(config any) error {
	configMap, ok := config.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid configuration format")
	}

	token, ok := configMap["credential_personal_access_token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("personal Access Token is required")
	}

	username, ok := configMap["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("username is required")
	}

	if patternsInterface, ok := configMap["repository_patterns"]; ok && patternsInterface != nil {
		patterns, ok := patternsInterface.([]any)
		if !ok {
			return fmt.Errorf("repository_patterns must be an array")
		}

		for i, p := range patterns {
			patternStr, ok := p.(string)
			if !ok {
				return fmt.Errorf("repository_patterns[%d] must be a string", i)
			}

			if err := validateRepositoryPattern(patternStr); err != nil {
				return fmt.Errorf("repository pattern at line %d: %w", i+1, err)
			}
		}
	}

	return nil
}

// validateRepositoryPattern validates a repository pattern format
func validateRepositoryPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Check if pattern contains '/'
	hasSlash := false
	for _, ch := range pattern {
		if ch == '/' {
			hasSlash = true
			break
		}
	}

	if !hasSlash {
		return fmt.Errorf("pattern must be in 'owner/repo' format (e.g., 'myorg/*', 'user/repo'). Got: '%s'", pattern)
	}

	// Split and validate parts
	parts := []string{}
	current := ""
	for _, ch := range pattern {
		if ch == '/' {
			if current == "" {
				return fmt.Errorf("invalid pattern: owner or repo part cannot be empty. Got: '%s'", pattern)
			}
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current == "" {
		return fmt.Errorf("invalid pattern: repo part cannot be empty. Got: '%s'", pattern)
	}
	parts = append(parts, current)

	// Should have exactly 2 parts (owner/repo)
	if len(parts) != 2 {
		return fmt.Errorf("pattern must have exactly one '/' separator (e.g., 'owner/repo'). Got: '%s'", pattern)
	}

	return nil
}
