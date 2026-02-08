package fetch

import "fmt"

type config struct {
	token        string
	userID       string
	targetDate   string
	workspaceURL string
}

func newConfig(cfg map[string]any, targetDate string) (*config, error) {
	token, ok := cfg["bot_token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("missing bot_token")
	}

	workspaceURL, ok := cfg["workspace_url"].(string)
	if !ok || workspaceURL == "" {
		return nil, fmt.Errorf("missing workspace_url")
	}

	userID, ok := cfg["user_id"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("missing user_id")
	}

	return &config{
		token:        token,
		userID:       userID,
		targetDate:   targetDate,
		workspaceURL: workspaceURL,
	}, nil
}
