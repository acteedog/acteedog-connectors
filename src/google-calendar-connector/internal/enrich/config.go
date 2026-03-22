package enrich

import (
	"encoding/json"
	"fmt"
)

// EnrichmentParams represents the enrichment_params extracted from context metadata
type EnrichmentParams struct {
	CalendarID string `json:"calendar_id"`
	EventID    string `json:"event_id"`
}

type config struct {
	contextType      string
	enrichmentParams EnrichmentParams
}

func newConfig(contextType string, params any) (*config, error) {
	pb, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enrichment_params: %w", err)
	}

	var enrichmentParams EnrichmentParams
	if err := json.Unmarshal(pb, &enrichmentParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal enrichment_params: %w", err)
	}

	return &config{
		contextType:      contextType,
		enrichmentParams: enrichmentParams,
	}, nil
}
