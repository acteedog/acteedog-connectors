package auth

import "fmt"

// Client is the interface for making authenticated HTTP GET requests to Google Calendar API
type Client interface {
	Get(url string) ([]byte, int, error)
}

// bearerAuthHeader returns the Authorization header value for Bearer token auth
func bearerAuthHeader(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
