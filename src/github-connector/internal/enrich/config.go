package enrich

import (
	"fmt"
)

type config struct {
	contextType      string
	token            string
	enrichmentParams map[string]any
}

func newConfig(contextType string, cfg map[string]any, params map[string]any) (*config, error) {
	token, ok := cfg["credential_personal_access_token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("missing personal access token")
	}

	return &config{
		contextType:      contextType,
		token:            token,
		enrichmentParams: params,
	}, nil
}
