// Note: run `go doc -all` in this package to see all of the types and functions available.
// ./pdk.gen.go contains the domain types from the host where your plugin will run.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"jira-connector/internal/core"

	"github.com/extism/go-pdk"
)

// connectorConfig represents the connector configuration provided by the host
type connectorConfig struct {
	CloudID       string   `json:"cloud_id"`
	Email         string   `json:"email"`
	APIToken      string   `json:"api_token"`
	ProjectIDs    []string `json:"project_ids"`
	SiteSubdomain string   `json:"site_subdomain"`
}

// parseConfig unmarshals the raw config into connectorConfig
func parseConfig(raw any) (*connectorConfig, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	var cfg connectorConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

// GetConfigSchema returns the configuration schema for the Jira connector
func GetConfigSchema() (ConfigSchema, error) {
	return ConfigSchema{
		Type: "object",
		Properties: map[string]any{
			"cloud_id": map[string]any{
				"type":        "string",
				"title":       "Cloud ID",
				"description": "Your Atlassian Cloud ID. You can find it at https://your-subdomain.atlassian.net/_edge/tenant_info",
			},
			"email": map[string]any{
				"type":        "string",
				"title":       "Email",
				"description": "Email address of your Jira account",
			},
			"api_token": map[string]any{
				"type":        "string",
				"title":       "API Token",
				"description": "Jira API Token for authentication with scope 'read:jira-user' and 'read:jira-work'",
			},
			"project_ids": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"title":       "Project IDs",
				"description": "List of Jira project IDs to fetch activities from (e.g., 10001) You can find it at https://your-subdomain.atlassian.net/rest/api/3/KEY",
			},
			"site_subdomain": map[string]any{
				"type":        "string",
				"title":       "Site Subdomain",
				"description": "Your Atlassian site subdomain (e.g., 'myorg' for myorg.atlassian.net)",
			},
		},
		Required: &[]string{
			"cloud_id",
			"email",
			"api_token",
			"project_ids",
			"site_subdomain",
		},
	}, nil
}

// TestConnection tests the Jira API connection using the provided configuration
func TestConnection(input TestConnectionRequest) error {
	pdk.Log(pdk.LogInfo, "TestConnection: Starting Jira API connection test")

	cfg, err := parseConfig(input.Config)
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Configuration parsing failed: %v", err))
		return err
	}

	if err := validateConfig(cfg); err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Configuration validation failed: %v", err))
		return err
	}

	url := fmt.Sprintf("%s/%s/rest/api/3/myself", core.JiraAPIBase, cfg.CloudID)

	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", basicAuthHeader(cfg.Email, cfg.APIToken))
	req.SetHeader("Accept", "application/json")

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Testing connection to: %s", url))

	res := req.Send()
	statusCode := res.Status()

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Response status: %d", statusCode))

	if statusCode == 200 {
		pdk.Log(pdk.LogInfo, "Connection test successful")
		return nil
	}

	var errorMsg string
	switch {
	case statusCode == 401:
		errorMsg = "Authentication failed: Invalid email or API Token"
	case statusCode == 403:
		errorMsg = "Access forbidden: Check token permissions"
	case statusCode >= 500:
		errorMsg = fmt.Sprintf("Jira API server error (status %d)", statusCode)
	default:
		errorMsg = fmt.Sprintf("Connection failed with status %d", statusCode)
	}

	var jiraError struct {
		Message       string   `json:"message"`
		ErrorMessages []string `json:"errorMessages"`
	}
	if err := json.Unmarshal(res.Body(), &jiraError); err == nil {
		if jiraError.Message != "" {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, jiraError.Message)
		} else if len(jiraError.ErrorMessages) > 0 {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, jiraError.ErrorMessages[0])
		}
	}

	pdk.Log(pdk.LogError, errorMsg)
	return fmt.Errorf("%s", errorMsg)
}

// validateConfig checks if required configuration fields are present and valid
func validateConfig(cfg *connectorConfig) error {
	if cfg.CloudID == "" {
		return fmt.Errorf("cloud_id is required")
	}
	if cfg.Email == "" {
		return fmt.Errorf("email is required")
	}
	if cfg.APIToken == "" {
		return fmt.Errorf("api_token is required")
	}
	if len(cfg.ProjectIDs) == 0 {
		return fmt.Errorf("project_ids must not be empty")
	}
	if cfg.SiteSubdomain == "" {
		return fmt.Errorf("site_subdomain is required")
	}
	return nil
}

// basicAuthHeader creates a Basic Auth header value from email and API token
func basicAuthHeader(email, apiToken string) string {
	credentials := email + ":" + apiToken
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encoded
}
