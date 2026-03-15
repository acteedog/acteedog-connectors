package main

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/core"
	"jira-connector/internal/match"

	pdk "github.com/extism/go-pdk"
)

// matchHTTPClient is the WASM-side implementation of match.HTTPClient using extism pdk.
type matchHTTPClient struct {
	email    string
	apiToken string
}

func (c *matchHTTPClient) FetchIssue(cloudID, _, _, issueKey string) (*match.IssueResponse, error) {
	url := fmt.Sprintf("%s/%s/rest/api/3/issue/%s?fields=project,parent,issuetype",
		core.JiraAPIBase, cloudID, issueKey)

	req := pdk.NewHTTPRequest(pdk.MethodGet, url)
	req.SetHeader("Authorization", basicAuthHeader(c.email, c.apiToken))
	req.SetHeader("Accept", "application/json")

	res := req.Send()
	if res.Status() != 200 {
		return nil, fmt.Errorf("Jira API error (status %d): %s", res.Status(), string(res.Body()))
	}

	var issue match.IssueResponse
	if err := json.Unmarshal(res.Body(), &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue response: %w", err)
	}

	return &issue, nil
}

// MatchContext matches the provided URLs against Jira browse URL patterns and returns context nodes.
// When a URL matches the configured site subdomain, the Jira API is called to resolve the
// issue hierarchy (source > project > [parent issue >] issue).
func MatchContext(input MatchContextRequest) (MatchContextResponse, error) {
	cfg, err := parseConfig(input.Config)
	if err != nil {
		return MatchContextResponse{}, fmt.Errorf("failed to parse config: %w", err)
	}

	httpClient := &matchHTTPClient{email: cfg.Email, apiToken: cfg.APIToken}
	matcher := match.NewContextMatcher(httpClient, cfg.CloudID, cfg.SiteSubdomain)

	results := make([]MatchContextResult, 0, len(input.Urls))

	for _, url := range input.Urls {
		coreContexts, err := matcher.MatchURLWithCredentials(url, cfg.Email, cfg.APIToken)
		if err != nil {
			// Log the error and return an empty result for this URL rather than aborting all
			pdk.Log(pdk.LogWarn, fmt.Sprintf("MatchContext: failed to match URL %s: %v", url, err))
			results = append(results, MatchContextResult{Url: url, Contexts: []Context{}})
			continue
		}

		contexts := make([]Context, 0, len(coreContexts))
		for _, c := range coreContexts {
			contexts = append(contexts, convertContext(c))
		}

		results = append(results, MatchContextResult{
			Url:      url,
			Contexts: contexts,
		})
	}

	return MatchContextResponse{Results: results}, nil
}
