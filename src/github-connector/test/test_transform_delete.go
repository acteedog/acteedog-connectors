package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("TransformEvent - DeleteEvent", func() {
		eventJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("TransformEvent", eventJSON)

		var activity map[string]any
		err := json.Unmarshal(outputJSON, &activity)
		xtptest.Assert("Activity can be parsed as JSON", err == nil, "Failed to parse activity JSON")
		if err != nil {
			return
		}

		// Verify basic activity fields
		xtptest.AssertEq("source is 'github'", activity["source"], "github")
		xtptest.AssertEq("activityType is 'delete'", activity["activityType"], "delete")

		// Verify activity
		xtptest.AssertEq("activityType is 'delete'", activity["activityType"], "delete")
		xtptest.AssertEq("Description is correct", activity["description"], "branch chore/golangci-lint was deleted")
		xtptest.AssertEq("Id is correct", activity["id"], "github:6031147775")
		xtptest.AssertEq("source is 'github'", activity["source"], "github")
		xtptest.AssertEq("Timestamp is correct", activity["timestamp"], "2025-11-18T00:37:03Z")
		xtptest.AssertEq("Title is correct", activity["title"], "Deleted branch chore/golangci-lint in ymtdzzz/otel-tui")
		xtptest.AssertEq("Url is correct", activity["url"], "https://github.com/ymtdzzz/otel-tui")

		// Verify metadata
		metadata, ok := activity["metadata"].(map[string]any)
		xtptest.Assert("metadata is a map", ok, "metadata is not a map")
		if ok {
			refType, _ := metadata["ref_type"].(string)
			xtptest.AssertEq("metadata has 'ref_type'", refType, "branch")

			ref, _ := metadata["ref"].(string)
			xtptest.AssertEq("metadata has 'ref'", ref, "chore/golangci-lint")

			deletedBy, _ := metadata["deleted_by"].(string)
			xtptest.AssertEq("metadata has 'deleted_by'", deletedBy, "ymtdzzz")

			pusherType, _ := metadata["pusher_type"].(string)
			xtptest.AssertEq("metadata has 'pusher_type'", pusherType, "user")
		}

		// Verify contexts
		contexts, ok := activity["contexts"].([]any)
		xtptest.Assert("contexts is an array", ok, "contexts is not an array")
		if ok {
			xtptest.AssertEq("has 2 contexts", len(contexts), 2)

			if len(contexts) >= 2 {
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
			}
		}
	})

	return 0
}

func main() {}
