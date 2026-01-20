package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("TransformEvent - IssueCommentEvent", func() {
		eventJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("TransformEvent", eventJSON)

		var activity map[string]any
		err := json.Unmarshal(outputJSON, &activity)
		xtptest.Assert("Activity can be parsed as JSON", err == nil, "Failed to parse activity JSON")
		if err != nil {
			return
		}

		// Verify activity
		xtptest.AssertEq("activityType is 'pr_comment'", activity["activityType"], "pr_comment")
		xtptest.AssertEq("Description is correct", activity["description"], "This is a comment on the PR.")
		xtptest.AssertEq("Id is correct", activity["id"], "github:4719356430")
		xtptest.AssertEq("source is 'github'", activity["source"], "github")
		xtptest.AssertEq("Timestamp is correct", activity["timestamp"], "2025-11-17T09:43:25Z")
		xtptest.AssertEq("Title is correct", activity["title"], "Commented on PR #52580")
		xtptest.AssertEq("Url is correct", activity["url"], "https://github.com/testorg/testrepo/pull/52580#issuecomment-3540831349")

		// Verify metadata
		metadata, ok := activity["metadata"].(map[string]any)
		xtptest.Assert("metadata is a map", ok, "metadata is not a map")
		if ok {
			commentID, _ := metadata["comment_id"].(float64)
			xtptest.AssertEq("metadata has 'comment_id'", int64(commentID), int64(3540831349))

			issueNumber, _ := metadata["issue_number"].(float64)
			xtptest.AssertEq("metadata has 'issue_number'", int(issueNumber), 52580)

			commentAuthor, _ := metadata["comment_author"].(string)
			xtptest.AssertEq("metadata has 'comment_author'", commentAuthor, "ymtdzzz")

			commentCreatedAt, _ := metadata["comment_created_at"].(string)
			xtptest.AssertEq("metadata has 'comment_created_at'", commentCreatedAt, "2025-11-17T09:43:25Z")
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
					xtptest.AssertEq("Id is 'github:repository:testorg/testrepo'", repoCtx["id"], "github:repository:testorg/testrepo")

					level, _ := repoCtx["level"].(float64)
					xtptest.AssertEq("Repository context level is 2", int(level), 2)

					xtptest.AssertEq("Name is correct", repoCtx["name"], "repository:testorg/testrepo")
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
							xtptest.AssertEq("repo is correct", repo, "testorg/testrepo")
						}
					}
				}

				prCtx, ok := contexts[2].(map[string]any)
				xtptest.Assert("PR context is a map", ok, "PR context is not a map")
				if ok {
					xtptest.AssertEq("ConnectorId is 'github'", prCtx["connectorId"], "github")
					xtptest.AssertEq("Id is 'github:pull_request:52580'", prCtx["id"], "github:pull_request:52580")

					level, _ := prCtx["level"].(float64)
					xtptest.AssertEq("PR context level is 3", int(level), 3)

					xtptest.AssertEq("Name is correct", prCtx["name"], "PR #52580")
					xtptest.AssertEq("ResourceType is 'pull_request'", prCtx["resourceType"], "pull_request")
					xtptest.AssertEq("ParentId is correct", prCtx["parentId"], "github:repository:testorg/testrepo")

					// Verify enrichment_params for PR
					metadata, ok := prCtx["metadata"].(map[string]any)
					xtptest.Assert("PR context has metadata", ok, "metadata not found")
					if ok {
						params, ok := metadata["enrichment_params"].(map[string]any)
						xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
						if ok {
							repo, _ := params["repo"]
							xtptest.AssertEq("repo is correct", repo, "testorg/testrepo")

							prNumber, _ := params["pr_number"].(string)
							xtptest.AssertEq("pr_number is correct", prNumber, "52580")
						}
					}
				}
			}
		}
	})

	return 0
}

func main() {}
