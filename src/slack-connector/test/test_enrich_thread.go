package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("ApplyThreadEnrichment", func() {
		apiResponseJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("ApplyThreadEnrichment", apiResponseJSON)

		var context map[string]any
		if err := json.Unmarshal(outputJSON, &context); err != nil {
			xtptest.AssertEq("Context can be parsed as JSON", err, nil)
			return
		}

		// Verify context
		xtptest.AssertEq("ConnectorId is 'slack'", context["connectorId"], "slack")
		xtptest.AssertEq("Description is parent message text", context["description"], "後からぶら下げる")
		xtptest.AssertEq("Id is 'slack:thread:C099VUEKVBN:1765613134.990399'", context["id"], "slack:thread:C099VUEKVBN:1765613134.990399")
		xtptest.AssertEq("Level is 3", int(context["level"].(float64)), 3)
		xtptest.AssertEq("Name is 'Thread 1765613134.990399'", context["name"], "Thread 1765613134.990399")
		xtptest.AssertEq("ParentId is 'slack:channel:C099VUEKVBN'", context["parentId"], "slack:channel:C099VUEKVBN")
		xtptest.AssertEq("ResourceType is 'thread'", context["resourceType"], "thread")
		xtptest.AssertEq("Title is correct", context["title"], "Thread: 後からぶら下げる")
		xtptest.AssertEq("URL is correctly formatted", context["url"], "https://test-workspace.slack.com/archives/C099VUEKVBN/p1765613134990399")
		xtptest.AssertEq("CreatedAt is correct", context["createdAt"], "2025-12-13T08:05:34.990399Z")
		xtptest.AssertEq("UpdatedAt is correct", context["createdAt"], "2025-12-13T08:05:34.990399Z")

		// Verify metadata
		metadata, ok := context["metadata"].(map[string]any)
		if !ok {
			xtptest.AssertEq("Metadata exists and is a map", ok, true)
			return
		}

		xtptest.AssertEq("Metadata.parent_user is 'U099SQHSJCW'", metadata["parent_user"], "U099SQHSJCW")
		xtptest.AssertEq("Metadata.parent_ts is '1765613134.990399'", metadata["parent_ts"], "1765613134.990399")
		xtptest.AssertEq("Metadata.thread_ts is '1765613134.990399'", metadata["thread_ts"], "1765613134.990399")
		xtptest.AssertEq("Metadata.team is 'T099VUE950C'", metadata["team"], "T099VUE950C")

		replyCount, _ := metadata["reply_count"].(float64)
		xtptest.AssertEq("Metadata.reply_count is 1", int64(replyCount), int64(1))

		replyUsersCount, _ := metadata["reply_users_count"].(float64)
		xtptest.AssertEq("Metadata.reply_users_count is 1", int64(replyUsersCount), int64(1))

		// Verify enrichment_params
		enrichmentParams, ok := metadata["enrichment_params"].(map[string]any)
		xtptest.AssertEq("Metadata.enrichment_params exists and is a map", ok, true)
		if !ok {
			return
		}
		xtptest.AssertEq("enrichment_params.channel_id is 'C099VUEKVBN'", enrichmentParams["channel_id"], "C099VUEKVBN")
		xtptest.AssertEq("enrichment_params.thread_ts is '1765613134.990399'", enrichmentParams["thread_ts"], "1765613134.990399")
	})

	return 0
}

func main() {}
