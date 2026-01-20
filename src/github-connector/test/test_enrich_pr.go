package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("ApplyPullRequestEnrichment", func() {
		apiResponseJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("ApplyPullRequestEnrichment", apiResponseJSON)

		var context map[string]any
		err := json.Unmarshal(outputJSON, &context)
		xtptest.Assert("Context can be parsed as JSON", err == nil, "Failed to parse context JSON")
		if err != nil {
			return
		}

		// Verify context
		xtptest.AssertEq("ConnectorId is 'github'", context["connectorId"], "github")
		xtptest.AssertEq("Description is enriched", context["description"], "This is a body of the PR.")
		xtptest.AssertEq("Id is correct", context["id"], "github:pull_request:52742")

		level, _ := context["level"].(float64)
		xtptest.AssertEq("Level is 3", int(level), 3)

		xtptest.AssertEq("Name is correct", context["name"], "PR #52742")
		xtptest.AssertEq("ParentId is correct", context["parentId"], "github:repository:testorg/testrepo")
		xtptest.AssertEq("ResourceType is 'pull_request'", context["resourceType"], "pull_request")
		xtptest.AssertEq("Title is enriched", context["title"], "Fix: remove all test cases")
		xtptest.AssertEq("Url is enriched", context["url"], "https://github.com/testorg/testrepo/pull/52742")
		xtptest.AssertEq("createdAt is enriched", context["createdAt"], "2025-11-11T00:52:36Z")
		xtptest.AssertEq("updatedAt is enriched", context["updatedAt"], "2025-11-13T05:34:50Z")

		// Verify metadata
		metadata, ok := context["metadata"].(map[string]any)
		xtptest.Assert("metadata is a map", ok, "metadata is not a map")
		if ok {
			xtptest.AssertEq("state is 'closed'", metadata["state"], "closed")
			xtptest.AssertEq("author is 'john'", metadata["author"], "john")

			assignees, ok := metadata["assignees"].([]any)
			xtptest.Assert("assignees is an array", ok, "assignees is not an array")
			if ok {
				xtptest.AssertEq("assignees has 1 item", len(assignees), 1)
				xtptest.AssertEq("assignee is 'john'", assignees[0], "john")
			}

			reviewers, ok := metadata["reviewers"].([]any)
			xtptest.Assert("reviewers is an array", ok, "reviewers is not an array")
			if ok {
				xtptest.AssertEq("reviewers has 2 items", len(reviewers), 2)
				xtptest.AssertEq("reviewer1 is correct", reviewers[0], "reviewer1")
				xtptest.AssertEq("reviewer2 is correct", reviewers[1], "reviewer2")
			}

			labels, ok := metadata["labels"].([]any)
			xtptest.Assert("labels is an array", ok, "labels is not an array")
			if ok {
				xtptest.AssertEq("labels has 2 items", len(labels), 2)
				xtptest.AssertEq("label1 is correct", labels[0], "label1")
				xtptest.AssertEq("label2 is correct", labels[1], "label2")
			}

			xtptest.AssertEq("base_branch is 'main'", metadata["base_branch"], "main")
			xtptest.AssertEq("head_branch is 'feature/awesome-branch'", metadata["head_branch"], "feature/awesome-branch")
			xtptest.AssertEq("milestone is empty", metadata["milestone"], "")

			additions, _ := metadata["additions"].(float64)
			xtptest.AssertEq("additions is 1", int(additions), 1)

			deletions, _ := metadata["deletions"].(float64)
			xtptest.AssertEq("deletions is 998", int(deletions), 998)

			changedFiles, _ := metadata["changed_files"].(float64)
			xtptest.AssertEq("changed_files is 29", int(changedFiles), 29)

			commits, _ := metadata["commits_count"].(float64)
			xtptest.AssertEq("commits_count is 1", int(commits), 1)

			merged, _ := metadata["merged"].(bool)
			xtptest.Assert("merged is true", merged, "merged should be true")

			xtptest.AssertEq("merged_at is correct", metadata["merged_at"], "2025-11-13T05:34:49Z")
			xtptest.AssertEq("merged_by is 'john'", metadata["merged_by"], "john")

			// Verify enrichment_params for pull request
			params, ok := metadata["enrichment_params"].(map[string]any)
			xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
			if ok {
				repo, _ := params["repo"]
				xtptest.AssertEq("repo is correct", repo, "testorg/testrepo")

				prNumber, _ := params["pr_number"].(string)
				xtptest.AssertEq("pr_number is correct", prNumber, "52742")
			}
		}
	})

	return 0
}

func main() {}
