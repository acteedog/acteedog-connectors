package auth

// Client is the interface for making authenticated HTTP GET requests.
// Concrete implementations handle token resolution based on auth method type.
type Client interface {
	// Get sends an authenticated GET request and returns the response body and status code.
	Get(url string) ([]byte, int, error)
}

// authorizationHeader builds the Authorization header value for GitHub API requests.
func authorizationHeader(token string) string { //nolint:unused // used in wasip1-only files (bearer.go, oauth.go) which are invisible to golangci-lint
	return "token " + token
}
