//go:build wasip1

// Note: run `go doc -all` in this package to see all of the types and functions available.
// ./pdk.gen.go contains the domain types from the host where your plugin will run.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"google-calendar-connector/internal/auth"
	"net/url"

	"github.com/extism/go-pdk"
)

// PKCE and state stored between BuildOAuthUrl and ExchangeOAuthCode calls.
// Package-level vars persist across WASM calls within the same plugin instance.
var (
	pkceCodeVerifier  string
	oauthState        string
	storedRedirectURI string
)

// GetConfigSchema returns the configuration schema for the Google Calendar connector
func GetConfigSchema() (ConfigSchema, error) {
	desc := "Authenticate via Google OAuth. Click Connect to open Google in your browser and grant calendar access."
	authMethods := []AuthMethod{
		{
			Id:          "oauth_web",
			Type:        AuthMethodTypeOauthWeb,
			Label:       "Google OAuth",
			Description: &desc,
			Fields:      []AuthField{},
		},
	}
	return ConfigSchema{
		Type: "object",
		Properties: map[string]any{
			"target_email": map[string]any{
				"type":        "string",
				"title":       "Email Address",
				"description": "Your Google account email address (used to filter calendars and events)",
			},
		},
		Required:    &[]string{"target_email"},
		AuthMethods: &authMethods,
	}, nil
}

func validateConfig(config any) error {
	configMap, ok := config.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid configuration format")
	}
	email, ok := configMap["target_email"].(string)
	if !ok || email == "" {
		return fmt.Errorf("target_email is required")
	}
	return nil
}

// BuildOAuthUrl builds the Google OAuth 2.0 authorization URL with PKCE.
func BuildOAuthUrl(input OAuthUrlRequest) (OAuthUrlResponse, error) {
	// Generate code verifier: 32 random bytes → base64url
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return OAuthUrlResponse{}, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	pkceCodeVerifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Compute code challenge: SHA256(verifier) → base64url
	h := sha256.Sum256([]byte(pkceCodeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

	// Generate state: 16 random bytes → hex
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return OAuthUrlResponse{}, fmt.Errorf("failed to generate state: %w", err)
	}
	oauthState = hex.EncodeToString(stateBytes)
	storedRedirectURI = input.RedirectUri

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", auth.GoogleCalendarClientID)
	params.Set("redirect_uri", input.RedirectUri)
	params.Set("scope", "https://www.googleapis.com/auth/calendar.readonly")
	params.Set("access_type", "offline")
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("state", oauthState)
	params.Set("prompt", "consent")

	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

	return OAuthUrlResponse{
		State: oauthState,
		Url:   authURL,
	}, nil
}

// ExchangeOAuthCode exchanges the authorization code for an access token.
func ExchangeOAuthCode(input OAuthCodeExchangeRequest) (OAuthTokenResponse, error) {
	const tokenURL = "https://oauth2.googleapis.com/token"

	body := url.Values{}
	body.Set("code", input.Code)
	body.Set("client_id", auth.GoogleCalendarClientID)
	body.Set("client_secret", auth.GoogleCalendarClientSecret)
	body.Set("redirect_uri", storedRedirectURI)
	body.Set("grant_type", "authorization_code")
	body.Set("code_verifier", pkceCodeVerifier)

	req := pdk.NewHTTPRequest(pdk.MethodPost, tokenURL)
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	req.SetHeader("User-Agent", "acteedog/google-calendar-connector")
	req.SetBody([]byte(body.Encode()))

	pdk.Log(pdk.LogInfo, "ExchangeOAuthCode: exchanging authorization code")
	res := req.Send()

	if res.Status() != 200 {
		return OAuthTokenResponse{}, fmt.Errorf("token exchange failed with status %d: %s", res.Status(), string(res.Body()))
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresIn    int64  `json:"expires_in,omitempty"`
		Error        string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		return OAuthTokenResponse{}, fmt.Errorf("failed to parse token response: %w", err)
	}
	if resp.Error != "" {
		return OAuthTokenResponse{}, fmt.Errorf("token exchange error: %s", resp.Error)
	}
	if resp.AccessToken == "" {
		return OAuthTokenResponse{}, fmt.Errorf("token exchange returned empty access_token")
	}

	pdk.Log(pdk.LogInfo, "ExchangeOAuthCode: token exchange successful")

	result := OAuthTokenResponse{AccessToken: resp.AccessToken}
	if resp.RefreshToken != "" {
		result.RefreshToken = &resp.RefreshToken
	}
	if resp.ExpiresIn != 0 {
		result.ExpiresIn = &resp.ExpiresIn
	}
	return result, nil
}

// StartDeviceFlow is not supported for Google Calendar (uses oauth_web only).
func StartDeviceFlow() (DeviceFlowResponse, error) {
	return DeviceFlowResponse{}, fmt.Errorf("device flow is not supported for Google Calendar; use oauth_web instead")
}

// PollDeviceToken is not supported for Google Calendar (uses oauth_web only).
func PollDeviceToken(input DeviceTokenRequest) (OAuthTokenResponse, error) {
	return OAuthTokenResponse{}, fmt.Errorf("device flow is not supported for Google Calendar; use oauth_web instead")
}

// TestConnection verifies the OAuth token works by calling the Calendar API.
func TestConnection(input TestConnectionRequest) error {
	pdk.Log(pdk.LogInfo, "TestConnection: testing Google Calendar API connection")

	if err := validateConfig(input.Config); err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("TestConnection: config validation failed: %v", err))
		return err
	}

	config, ok := input.Config.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid configuration format")
	}

	client, err := auth.NewClient(config)
	if err != nil {
		return err
	}

	url := "https://www.googleapis.com/calendar/v3/users/me/calendarList?maxResults=1"
	body, statusCode, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("TestConnection: response status %d", statusCode))

	if statusCode == 200 {
		pdk.Log(pdk.LogInfo, "TestConnection: connection successful")
		return nil
	}

	// Try to parse Google's error format
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error.Message != "" {
		return fmt.Errorf("Google Calendar API error (%d): %s", errResp.Error.Code, errResp.Error.Message)
	}

	return fmt.Errorf("Google Calendar API request failed with status %d", statusCode)
}
