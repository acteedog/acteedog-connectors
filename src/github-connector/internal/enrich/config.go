package enrich

type config struct {
	contextType      string
	enrichmentParams map[string]any
}

func newConfig(contextType string, cfg map[string]any, params map[string]any) (*config, error) {
	return &config{
		contextType:      contextType,
		enrichmentParams: params,
	}, nil
}
