// Note: run `go doc -all` in this package to see all of the types and functions available.
// ./pdk.gen.go contains the domain types from the host where your plugin will run.
package main

import (
	"encoding/json"
	"fmt"
	"slack-connector/internal/core"

	"github.com/extism/go-pdk"
)

// GetConfigSchema returns the configuration schema for the Slack connector
func GetConfigSchema() (ConfigSchema, error) {
	return ConfigSchema{
		Type: "object",
		Properties: map[string]any{
			"bot_token": map[string]any{
				"type":        "string",
				"title":       "Bot User OAuth Token",
				"description": "Slack Bot Token (xoxb-...) with search:read scope",
			},
			"user_id": map[string]any{
				"type":        "string",
				"title":       "User ID",
				"description": "Slack User ID to fetch messages for (e.g., U1234567890)",
			},
			"workspace_url": map[string]any{
				"type":        "string",
				"title":       "Workspace URL",
				"description": "Your Slack workspace domain (e.g., your-workspace.slack.com)",
				"placeholder": "your-workspace.slack.com",
			},
		},
		Required: &[]string{
			"bot_token",
			"user_id",
			"workspace_url",
		},
	}, nil
}

// TestConnection tests the Slack API connection using the provided configuration
func TestConnection(input TestConnectionRequest) error {
	pdk.Log(pdk.LogInfo, "TestConnection: Starting Slack API connection test")

	err := validateConfig(input.Config)
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Configuration validation failed: %v", err))
		return err
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid configuration format")
	}

	botToken, ok := config["bot_token"].(string)
	if !ok || botToken == "" {
		return fmt.Errorf("bot_token is required")
	}

	url := fmt.Sprintf("%s/auth.test", core.SlackAPIBaseURL)

	req := pdk.NewHTTPRequest(pdk.MethodPost, url)
	req.SetHeader("Authorization", fmt.Sprintf("Bearer %s", botToken))
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Testing connection to: %s", url))

	res := req.Send()
	statusCode := res.Status()

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Response status: %d", statusCode))

	var authResponse struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
		Team  string `json:"team"`
	}

	if err := json.Unmarshal(res.Body(), &authResponse); err != nil {
		errorMsg := fmt.Sprintf("Failed to parse Slack API response: %v", err)
		pdk.Log(pdk.LogError, errorMsg)
		return fmt.Errorf("%s", errorMsg)
	}

	if authResponse.Ok {
		pdk.Log(pdk.LogInfo, "Connection test successful")
		return nil
	}

	var errorMsg string
	switch authResponse.Error {
	case "invalid_auth":
		errorMsg = "Authentication failed: Invalid or expired Bot Token"
	case "account_inactive":
		errorMsg = "Authentication failed: Account is inactive"
	case "token_revoked":
		errorMsg = "Authentication failed: Token has been revoked"
	case "no_permission":
		errorMsg = "Authentication failed: Token lacks required permissions"
	default:
		errorMsg = fmt.Sprintf("Connection failed: %s", authResponse.Error)
	}

	pdk.Log(pdk.LogError, errorMsg)

	return fmt.Errorf("%s", errorMsg)
}

// validateConfig checks if required configuration fields are present
func validateConfig(config any) error {
	configMap, ok := config.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid configuration format")
	}

	botToken, ok := configMap["bot_token"].(string)
	if !ok || botToken == "" {
		return fmt.Errorf("bot_token is required")
	}

	userID, ok := configMap["user_id"].(string)
	if !ok || userID == "" {
		return fmt.Errorf("user_id is required")
	}

	workspaceURL, ok := configMap["workspace_url"].(string)
	if !ok || workspaceURL == "" {
		return fmt.Errorf("workspace_url is required")
	}

	return nil
}

// getStringValue safely extracts string value from map
func getStringValue(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
