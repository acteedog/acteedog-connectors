//go:build wasip1

package auth

// NewClient creates an appropriate Client based on the active_auth_method ID in cfg.
// Google Calendar only supports "oauth_web".
func NewClient(cfg map[string]any) (Client, error) {
	return newOAuthClient(cfg)
}
