package enrich

import "fmt"

type config struct {
	contextType      string
	token            string
	workspaceURL     string
	enrichmentParams map[string]any
}

func newConfig(contextType string, cfg map[string]any, params map[string]any) (*config, error) {
	token, ok := cfg["bot_token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("missing bot_token")
	}

	workspaceURL, ok := cfg["workspace_url"].(string)
	if !ok || workspaceURL == "" {
		return nil, fmt.Errorf("missing workspace_url")
	}

	return &config{
		contextType:      contextType,
		token:            token,
		workspaceURL:     workspaceURL,
		enrichmentParams: params,
	}, nil
}
