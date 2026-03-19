//go:build wasip1

package auth

// NewClient creates an appropriate Client based on the active_auth_method ID in cfg.
// The auth method ID is defined by the connector's GetConfigSchema:
//   - "token"        → BearerClient  (uses cfg["personal_access_token"])
//   - "oauth_device" → OAuthClient   (uses cfg["oauth_access_token"])
func NewClient(cfg map[string]any) (Client, error) {
	method, _ := cfg["active_auth_method"].(string)
	switch method {
	case "oauth_device", "oauth_web":
		return newOAuthClient(cfg)
	default:
		// "token" or unset → PAT bearer auth
		return newBearerClient(cfg)
	}
}
