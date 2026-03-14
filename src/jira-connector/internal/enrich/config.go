package enrich

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/core"
)

// EnrichmentParams represents the enrichment_params extracted from context metadata
type EnrichmentParams struct {
	ProjectID string `json:"project_id"`
	IssueID   string `json:"issue_id"`
}

type config struct {
	*core.ConnectorConfig
	contextType      string
	enrichmentParams EnrichmentParams
}

func newConfig(contextType string, cfg any, params any) (*config, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var connCfg core.ConnectorConfig
	if err := json.Unmarshal(b, &connCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := connCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid connector config: %w", err)
	}

	pb, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enrichment_params: %w", err)
	}

	var enrichmentParams EnrichmentParams
	if err := json.Unmarshal(pb, &enrichmentParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal enrichment_params: %w", err)
	}

	return &config{
		ConnectorConfig:  &connCfg,
		contextType:      contextType,
		enrichmentParams: enrichmentParams,
	}, nil
}
