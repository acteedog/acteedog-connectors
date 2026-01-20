package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("TransformEvent - IssuesEvent", func() {
		eventJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("TransformEvent", eventJSON)

		var activity map[string]any
		err := json.Unmarshal(outputJSON, &activity)
		xtptest.Assert("Activity can be parsed as JSON", err == nil, "Failed to parse activity JSON")
		if err != nil {
			return
		}

		// Verify activity
		xtptest.AssertEq("activityType is 'issues'", activity["activityType"], "issues")
		xtptest.AssertEq("Description is correct", activity["description"], "Issue #375 was labeled")
		xtptest.AssertEq("Id is correct", activity["id"], "github:4716363699")
		xtptest.AssertEq("source is 'github'", activity["source"], "github")
		xtptest.AssertEq("Timestamp is correct", activity["timestamp"], "2025-11-17T07:15:21Z")
		xtptest.AssertEq("Title is correct", activity["title"], "Issue #375 labeled in ymtdzzz/otel-tui")
		xtptest.AssertEq("Url is correct", activity["url"], "https://github.com/ymtdzzz/otel-tui/issues/375")

		// Verify metadata
		metadata, ok := activity["metadata"].(map[string]any)
		xtptest.Assert("metadata is a map", ok, "metadata is not a map")
		if ok {
			issueNumber, _ := metadata["issue_number"].(float64)
			xtptest.AssertEq("metadata has 'issue_number'", int(issueNumber), 375)

			action, _ := metadata["action"].(string)
			xtptest.AssertEq("metadata has 'action'", action, "labeled")

			state, _ := metadata["state"].(string)
			xtptest.AssertEq("metadata has 'state'", state, "open")

			labels, ok := metadata["labels"].([]any)
			xtptest.Assert("metadata has 'labels' array", ok, "labels is not an array")
			xtptest.AssertEq("labels has 1 label", len(labels), 1)
			xtptest.AssertEq("label name is 'CI/CD'", labels[0], "CI/CD")
		}

		// Verify contexts
		contexts, ok := activity["contexts"].([]any)
		xtptest.Assert("contexts is an array", ok, "contexts is not an array")
		if ok {
			xtptest.AssertEq("has 3 contexts", len(contexts), 3)

			if len(contexts) >= 3 {
				sourceCtx, ok := contexts[0].(map[string]any)
				xtptest.Assert("source context is a map", ok, "source context is not a map")
				if ok {
					xtptest.AssertEq("ConnectorId is 'github'", sourceCtx["connectorId"], "github")
					xtptest.AssertEq("Id is 'github:source'", sourceCtx["id"], "github:source")

					level, _ := sourceCtx["level"].(float64)
					xtptest.AssertEq("Source context level is 1", int(level), 1)

					xtptest.AssertEq("Name is correct", sourceCtx["name"], "github:source")
					xtptest.AssertEq("ParentId is empty", sourceCtx["parentId"], "")
					xtptest.AssertEq("ResourceType is 'source'", sourceCtx["resourceType"], "source")

					metadata, ok := sourceCtx["metadata"].(map[string]any)
					xtptest.Assert("Source context has metadata", ok, "metadata not found")
					if ok {
						_, ok := metadata["enrichment_params"].(map[string]any)
						xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
					}
				}

				repoCtx, ok := contexts[1].(map[string]any)
				xtptest.Assert("Repository context is a map", ok, "Repository context is not a map")
				if ok {
					xtptest.AssertEq("ConnectorId is 'github'", repoCtx["connectorId"], "github")
					xtptest.AssertEq("Id is 'github:repository:ymtdzzz/otel-tui'", repoCtx["id"], "github:repository:ymtdzzz/otel-tui")

					level, _ := repoCtx["level"].(float64)
					xtptest.AssertEq("Repository context level is 2", int(level), 2)

					xtptest.AssertEq("Name is correct", repoCtx["name"], "repository:ymtdzzz/otel-tui")
					xtptest.AssertEq("ParentId is correct", repoCtx["parentId"], "github:source")
					xtptest.AssertEq("ResourceType is 'repository'", repoCtx["resourceType"], "repository")

					// Verify enrichment_params for repository
					metadata, ok := repoCtx["metadata"].(map[string]any)
					xtptest.Assert("Repository context has metadata", ok, "metadata not found")
					if ok {
						params, ok := metadata["enrichment_params"].(map[string]any)
						xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
						if ok {
							repo, _ := params["repo"]
							xtptest.AssertEq("repo is correct", repo, "ymtdzzz/otel-tui")
						}
					}
				}

				issueCtx, ok := contexts[2].(map[string]any)
				xtptest.Assert("Issue context is a map", ok, "Issue context is not a map")
				if ok {
					xtptest.AssertEq("ConnectorId is 'github'", issueCtx["connectorId"], "github")
					xtptest.AssertEq("Id is 'github:issue:375'", issueCtx["id"], "github:issue:375")

					level, _ := issueCtx["level"].(float64)
					xtptest.AssertEq("Issue context level is 3", int(level), 3)

					xtptest.AssertEq("Name is correct", issueCtx["name"], "Issue #375")
					xtptest.AssertEq("ParentId is correct", issueCtx["parentId"], "github:repository:ymtdzzz/otel-tui")
					xtptest.AssertEq("ResourceType is 'issue'", issueCtx["resourceType"], "issue")

					// Verify enrichment_params for PR
					metadata, ok := issueCtx["metadata"].(map[string]any)
					xtptest.Assert("Issue context has metadata", ok, "metadata not found")
					if ok {
						params, ok := metadata["enrichment_params"].(map[string]any)
						xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
						if ok {
							repo, _ := params["repo"]
							xtptest.AssertEq("repo is correct", repo, "ymtdzzz/otel-tui")

							issueNumber, _ := params["issue_number"].(string)
							xtptest.AssertEq("issue_number is correct", issueNumber, "375")
						}
					}
				}
			}
		}
	})

	return 0
}

func main() {}
