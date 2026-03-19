// Note: run `go doc -all` in this package to see all of the types and functions available.
// ./pdk.gen.go contains the domain types from the host where your plugin will run.
package main

import (
	"encoding/json"
	"fmt"
	"github-connector/internal/auth"
	"github-connector/internal/core"

	"github.com/extism/go-pdk"
)

// GetConfigSchema returns the configuration schema for the GitHub connector
func GetConfigSchema() (ConfigSchema, error) {
	secretTrue := true
	authMethods := []AuthMethod{
		{
			Id:          "token",
			Type:        AuthMethodTypeBearer,
			Label:       "Personal Access Token",
			Description: strPtr("GitHub Personal Access Token authentication. Required scopes: repo, read:user"),
			Fields: []AuthField{
				{Key: "personal_access_token", Name: "Personal Access Token", Secret: &secretTrue},
			},
		},
		{
			Id:          "oauth_device",
			Type:        AuthMethodTypeOauthDevice,
			Label:       "GitHub App (Device Flow)",
			Description: strPtr("Authenticate via GitHub Device Flow. Open GitHub in browser and enter a code. No redirect required."),
			Fields:      []AuthField{},
		},
	}
	return ConfigSchema{
		Type: "object",
		Properties: map[string]any{
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
		Required:    &[]string{"username"},
		AuthMethods: &authMethods,
	}, nil
}

func strPtr(s string) *string { return &s }

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

	authClient, err := auth.NewClient(config)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/user", core.GithubAPIBaseURL)

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Testing connection to: %s", url))

	body, statusCode, err := authClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Response status: %d", statusCode))

	if statusCode == 200 {
		pdk.Log(pdk.LogInfo, "Connection test successful")
		return nil
	}

	var errorMsg string
	if statusCode == 401 {
		errorMsg = "Authentication failed: Invalid or expired token"
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
	if err := json.Unmarshal(body, &githubError); err == nil && githubError.Message != "" {
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

// BuildOAuthUrl is a stub for the OAuth Web Flow.
// GitHub App requires client_secret for token exchange, which cannot be safely
// embedded in a desktop app binary. Use oauth_device (Device Flow) instead.
func BuildOAuthUrl(input OAuthUrlRequest) (OAuthUrlResponse, error) {
	return OAuthUrlResponse{}, fmt.Errorf("oauth_web flow is not supported for GitHub App; use oauth_device (Device Flow) instead")
}

// ExchangeOAuthCode is a stub for the OAuth Web Flow.
// See BuildOAuthUrl for the reason.
func ExchangeOAuthCode(input OAuthCodeExchangeRequest) (OAuthTokenResponse, error) {
	return OAuthTokenResponse{}, fmt.Errorf("oauth_web flow is not supported for GitHub App; use oauth_device (Device Flow) instead")
}

// StartDeviceFlow initiates the GitHub OAuth Device Flow.
// It requests a device code and user code from GitHub, which the host then
// displays to the user. The user visits verification_uri and enters user_code.
func StartDeviceFlow() (DeviceFlowResponse, error) {
	body := fmt.Sprintf("client_id=%s&scope=repo,read:user", auth.GithubAppClientID)

	req := pdk.NewHTTPRequest(pdk.MethodPost, "https://github.com/login/device/code")
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	req.SetHeader("User-Agent", "acteedog-github-connector")
	req.SetBody([]byte(body))

	pdk.Log(pdk.LogInfo, "StartDeviceFlow: requesting device code")
	res := req.Send()

	if res.Status() != 200 {
		return DeviceFlowResponse{}, fmt.Errorf("device flow request failed with status %d", res.Status())
	}

	var resp struct {
		DeviceCode       string `json:"device_code"`
		UserCode         string `json:"user_code"`
		VerificationUri  string `json:"verification_uri"`
		ExpiresIn        int64  `json:"expires_in"`
		Interval         int64  `json:"interval"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		return DeviceFlowResponse{}, fmt.Errorf("failed to parse device code response: %w", err)
	}
	if resp.Error != "" {
		return DeviceFlowResponse{}, fmt.Errorf("device flow error: %s: %s", resp.Error, resp.ErrorDescription)
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("StartDeviceFlow: user_code=%s verification_uri=%s", resp.UserCode, resp.VerificationUri))
	return DeviceFlowResponse{
		DeviceCode:      resp.DeviceCode,
		UserCode:        resp.UserCode,
		VerificationUri: resp.VerificationUri,
		ExpiresIn:       resp.ExpiresIn,
		Interval:        resp.Interval,
	}, nil
}

// PollDeviceToken polls the GitHub token endpoint for a Device Flow token.
// Returns an OAuthTokenResponse with an empty AccessToken when authorization
// is still pending (the host should continue polling at the specified interval).
// Returns an error for terminal states: expired_token, access_denied, etc.
func PollDeviceToken(input DeviceTokenRequest) (OAuthTokenResponse, error) {
	body := fmt.Sprintf(
		"client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code",
		auth.GithubAppClientID,
		input.DeviceCode,
	)

	req := pdk.NewHTTPRequest(pdk.MethodPost, "https://github.com/login/oauth/access_token")
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	req.SetHeader("User-Agent", "acteedog-github-connector")
	req.SetBody([]byte(body))

	res := req.Send()

	if res.Status() != 200 {
		return OAuthTokenResponse{}, fmt.Errorf("token poll failed with status %d", res.Status())
	}

	var resp struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token,omitempty"`
		ExpiresIn        int64  `json:"expires_in,omitempty"`
		Error            string `json:"error,omitempty"`
		ErrorDescription string `json:"error_description,omitempty"`
	}
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		return OAuthTokenResponse{}, fmt.Errorf("failed to parse token poll response: %w", err)
	}

	switch resp.Error {
	case "":
		// Success or access_token is present
	case "authorization_pending", "slow_down":
		// Not yet authorized — return empty token to signal the host to keep polling
		pdk.Log(pdk.LogDebug, fmt.Sprintf("PollDeviceToken: %s, continuing poll", resp.Error))
		return OAuthTokenResponse{AccessToken: ""}, nil
	default:
		// Terminal error: expired_token, access_denied, incorrect_device_code, etc.
		return OAuthTokenResponse{}, fmt.Errorf("device flow error: %s: %s", resp.Error, resp.ErrorDescription)
	}

	if resp.AccessToken == "" {
		return OAuthTokenResponse{}, fmt.Errorf("token poll returned empty access token without error")
	}

	pdk.Log(pdk.LogInfo, "PollDeviceToken: token acquired successfully")
	result := OAuthTokenResponse{AccessToken: resp.AccessToken}
	if resp.RefreshToken != "" {
		result.RefreshToken = &resp.RefreshToken
	}
	if resp.ExpiresIn != 0 {
		result.ExpiresIn = &resp.ExpiresIn
	}
	return result, nil
}
