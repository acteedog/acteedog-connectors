//go:build wasip1

package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/extism/go-pdk"
)

// GithubAppClientID is injected at build time via:
//
//	-ldflags "-X github-connector/internal/auth.GithubAppClientID=<value>"
//
// Set GITHUB_APP_CLIENT_ID env var before building (see xtp.toml).
var GithubAppClientID string

// oauthClient authenticates using an OAuth access token (Device Flow / Web Flow).
// When a request returns HTTP 401, it transparently attempts to refresh the
// access token using the stored refresh_token before retrying once.
type oauthClient struct {
	accessToken  string
	refreshToken string // empty if not available
}

func newOAuthClient(cfg map[string]any) (*oauthClient, error) {
	token, ok := cfg["oauth_access_token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("not connected via OAuth: please connect via GitHub App (Device Flow) first")
	}
	refreshToken, _ := cfg["oauth_refresh_token"].(string)
	return &oauthClient{
		accessToken:  token,
		refreshToken: refreshToken,
	}, nil
}

func (c *oauthClient) Get(url string) ([]byte, int, error) {
	body, status, err := c.doRequest(url)
	if err != nil {
		return nil, status, err
	}

	// On 401: attempt a transparent token refresh and retry once.
	if status == 401 && c.refreshToken != "" {
		pdk.Log(pdk.LogInfo, "Refreshing token...")
		if refreshErr := c.refresh(); refreshErr != nil {
			// Refresh failed – surface a clear re-auth message.
			return nil, status, fmt.Errorf(
				"OAuth token expired and refresh failed: %w – please reconnect via GitHub App (Device Flow)",
				refreshErr,
			)
		}
		pdk.Log(pdk.LogInfo, "Refresh token completed")
		// Retry with the new access token.
		body, status, err = c.doRequest(url)
	}

	return body, status, err
}

// doRequest performs a single GET request with the current access token.
func (c *oauthClient) doRequest(url string) ([]byte, int, error) {
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", authorizationHeader(c.accessToken))
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/github-connector")
	res := req.Send()
	return res.Body(), int(res.Status()), nil
}

// refresh exchanges the stored refresh_token for a new access_token via the
// GitHub token endpoint, updates c.accessToken / c.refreshToken in-memory,
// and persists the new tokens to the host keychain via the store_oauth_tokens
// host function.
func (c *oauthClient) refresh() error {
	const tokenURL = "https://github.com/login/oauth/access_token"

	body := fmt.Sprintf(
		"client_id=%s&grant_type=refresh_token&refresh_token=%s",
		GithubAppClientID,
		c.refreshToken,
	)

	req := pdk.NewHTTPRequest(pdk.MethodPost, tokenURL)
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	req.SetHeader("User-Agent", "acteedog/github-connector")
	req.SetBody([]byte(body))
	res := req.Send()

	if res.Status() != 200 {
		return fmt.Errorf("token refresh request failed with status %d", res.Status())
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Error        string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		return fmt.Errorf("failed to parse token refresh response: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("token refresh error: %s", resp.Error)
	}
	if resp.AccessToken == "" {
		return fmt.Errorf("token refresh returned empty access_token")
	}

	// Update in-memory tokens.
	c.accessToken = resp.AccessToken
	if resp.RefreshToken != "" {
		c.refreshToken = resp.RefreshToken
	}

	// Persist to host keychain.
	payload := map[string]string{"access_token": c.accessToken}
	if c.refreshToken != "" {
		payload["refresh_token"] = c.refreshToken
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal token payload: %w", err)
	}
	storeOAuthTokens(strings.TrimSpace(string(jsonBytes)))

	return nil
}
