package main

import (
	"github-connector/internal/core"
	"github-connector/internal/match"
)

// MatchContext matches the provided URLs against GitHub URL patterns and returns context nodes.
// No external API calls are made; context hierarchy is constructed from URL captures alone.
func MatchContext(input MatchContextRequest) (MatchContextResponse, error) {
	gen := core.NewContextGenerator()
	results := make([]MatchContextResult, 0, len(input.Urls))

	for _, url := range input.Urls {
		coreContexts := match.MatchURL(gen, url)
		contexts := make([]Context, 0, len(coreContexts))
		for _, c := range coreContexts {
			contexts = append(contexts, convertCoreContext(c))
		}
		results = append(results, MatchContextResult{
			Url:      url,
			Contexts: contexts,
		})
	}

	return MatchContextResponse{Results: results}, nil
}

// convertCoreContext converts a core.Context to the pdk-generated Context type.
func convertCoreContext(c *core.Context) Context {
	return Context{
		Id:           c.Id,
		Name:         c.Name,
		ParentId:     c.ParentId,
		ConnectorId:  c.ConnectorId,
		ResourceType: c.ResourceType,
		Title:        c.Title,
		Description:  c.Description,
		Url:          c.Url,
		Metadata:     c.Metadata,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
