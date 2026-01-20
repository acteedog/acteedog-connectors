package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("ApplyIssueEnrichment", func() {
		apiResponseJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("ApplyIssueEnrichment", apiResponseJSON)

		var context map[string]any
		err := json.Unmarshal(outputJSON, &context)
		xtptest.Assert("Context can be parsed as JSON", err == nil, "Failed to parse context JSON")
		if err != nil {
			return
		}

		// Verify context
		xtptest.AssertEq("ConnectorId is 'github'", context["connectorId"], "github")
		xtptest.AssertEq("Description is enriched", context["description"], "It would be great if you could use `j` and `k` on the \"Trace Timeline\" page as well. Currently, it works on the \"main\" trace page, but not in the timeline.")
		xtptest.AssertEq("Id is correct", context["id"], "github:issue:340")

		level, _ := context["level"].(float64)
		xtptest.AssertEq("Level is 3", int(level), 3)

		xtptest.AssertEq("Name is correct", context["name"], "Issue #340")
		xtptest.AssertEq("ParentId is correct", context["parentId"], "github:repository:ymtdzzz/otel-tui")
		xtptest.AssertEq("ResourceType is 'issue'", context["resourceType"], "issue")
		xtptest.AssertEq("Title is enriched", context["title"], "Use `j` and `k` for navigation in trace timeline")
		xtptest.AssertEq("Url is enriched", context["url"], "https://github.com/ymtdzzz/otel-tui/issues/340")
		xtptest.AssertEq("createdAt is enriched", context["createdAt"], "2025-10-06T19:55:39Z")
		xtptest.AssertEq("updatedAt is enriched", context["updatedAt"], "2025-11-08T07:32:14Z")

		// Verify metadata
		metadata, ok := context["metadata"].(map[string]any)
		xtptest.Assert("metadata is a map", ok, "metadata is not a map")
		if ok {
			xtptest.AssertEq("state is 'open'", metadata["state"], "open")
			xtptest.AssertEq("author is 'testuser'", metadata["author"], "testuser")

			assignees, ok := metadata["assignees"].([]any)
			xtptest.Assert("assignees is an array", ok, "assignees is not an array")
			if ok {
				xtptest.AssertEq("assignees is empty", len(assignees), 0)
			}

			labels, ok := metadata["labels"].([]any)
			xtptest.Assert("labels is an array", ok, "labels is not an array")
			if ok {
				xtptest.AssertEq("labels has 5 items", len(labels), 5)
				xtptest.AssertEq("label1 is 'bug'", labels[0], "bug")
				xtptest.AssertEq("label2 is 'enhancement'", labels[1], "enhancement")
				xtptest.AssertEq("label3 is 'good first issue'", labels[2], "good first issue")
				xtptest.AssertEq("label4 is 'UI'", labels[3], "UI")
				xtptest.AssertEq("label5 is 'signal: traces'", labels[4], "signal: traces")
			}

			xtptest.AssertEq("milestone is empty", metadata["milestone"], "")

			comments, _ := metadata["comments"].(float64)
			xtptest.AssertEq("comments is 2", int(comments), 2)

			// Verify enrichment_params for issue
			params, ok := metadata["enrichment_params"].(map[string]any)
			xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
			if ok {
				repo, _ := params["repo"]
				xtptest.AssertEq("repo is correct", repo, "ymtdzzz/otel-tui")

				issueNumber, _ := params["issue_number"].(string)
				xtptest.AssertEq("issue_number is correct", issueNumber, "340")
			}
		}
	})

	return 0
}

func main() {}
