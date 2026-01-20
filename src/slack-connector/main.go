// Note: run `go doc -all` in this package to see all of the types and functions available.
// ./pdk.gen.go contains the domain types from the host where your plugin will run.
package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/extism/go-pdk"
)

const (
	// ConnectorVersion is the version of this connector
	ConnectorVersion = "0.1.0"
	// ConnectorID is the unique identifier for this connector
	ConnectorID = "slack"
	// SlackAPIBaseURL is the base URL for Slack API
	SlackAPIBaseURL = "https://slack.com/api"
)

// Resource type constants for context identification
const (
	ResourceTypeSource  = "source"
	ResourceTypeChannel = "channel"
	ResourceTypeThread  = "thread"
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

	url := fmt.Sprintf("%s/auth.test", SlackAPIBaseURL)

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

// getNestedString safely extracts nested string value
func getNestedString(m map[string]any, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - extract string
			return getStringValue(current, key)
		}
		// Navigate deeper
		if nested, ok := current[key].(map[string]any); ok {
			current = nested
		} else {
			return ""
		}
	}
	return ""
}

// getIntValue safely extracts int64 value from map
func getIntValue(m map[string]any, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	return 0
}

// getBoolValue safely extracts bool value from map
func getBoolValue(m map[string]any, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// parseThreadTS extracts thread_ts from permalink
// Example: https://...slack.com/archives/C099VUEKVBN/p1765613227980829?thread_ts=1765613134.990399
// Returns: "1765613134.990399" or empty string if not found
func parseThreadTS(permalink string) string {
	re := regexp.MustCompile(`thread_ts=([0-9.]+)`)
	matches := re.FindStringSubmatch(permalink)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// formatSlackTS converts Slack timestamp to URL format
// Example: "1765611321.248519" -> "1765611321248519"
func formatSlackTS(ts string) string {
	return strings.ReplaceAll(ts, ".", "")
}

// convertSlackTSToTime converts Slack timestamp to time.Time
// Slack timestamp format: "1765611321.248519" (epoch seconds with microseconds)
func convertSlackTSToTime(ts string) (time.Time, error) {
	// Parse as float to handle the decimal part
	epochFloat, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Convert to Unix timestamp (seconds and nanoseconds)
	seconds := int64(epochFloat)
	nanos := int64((epochFloat - float64(seconds)) * 1e9)

	return time.Unix(seconds, nanos), nil
}

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}
