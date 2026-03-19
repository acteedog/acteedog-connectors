//go:build wasip1

package auth

import (
	"fmt"

	"github.com/extism/go-pdk"
)

// bearerClient authenticates using a Personal Access Token (PAT).
type bearerClient struct {
	token string
}

func newBearerClient(cfg map[string]any) (*bearerClient, error) {
	token, ok := cfg["personal_access_token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("personal access token is required")
	}
	return &bearerClient{token: token}, nil
}

func (c *bearerClient) Get(url string) ([]byte, int, error) {
	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", authorizationHeader(c.token))
	req.SetHeader("Accept", "application/vnd.github+json")
	req.SetHeader("User-Agent", "acteedog/github-connector")
	res := req.Send()
	return res.Body(), int(res.Status()), nil
}
