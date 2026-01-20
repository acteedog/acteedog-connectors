package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("TransformMessage (Standalone without thread_ts)", func() {
		messageJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("TransformMessage", messageJSON)

		var activity map[string]any
		if err := json.Unmarshal(outputJSON, &activity); err != nil {
			xtptest.AssertEq("Activity can be parsed as JSON", err, nil)
			return
		}

		// Verify activity
		xtptest.AssertEq("ActivityType is 'message'", activity["activityType"], "message")
		xtptest.AssertEq("Description is message text", activity["description"], "単体メッセージ")
		xtptest.AssertEq("ID is message timestamp", activity["id"], "slack:1765611321.248519")
		xtptest.AssertEq("Source is 'Slack'", activity["source"], "Slack")
		xtptest.AssertEq("Timestamp is correct", activity["timestamp"], "2025-12-13T07:35:21Z")
		xtptest.AssertEq("Title contains channel name", activity["title"], "Message in #general")
		xtptest.AssertEq("URL is permalink without thread_ts", activity["url"], "https://test-workspace.slack.com/archives/C099VUEKVBN/p1765611321248519")

		// Verify metadata
		metadata, ok := activity["metadata"].(map[string]any)
		xtptest.AssertEq("Metadata exists and is a map", ok, true)
		if !ok {
			return
		}

		xtptest.AssertEq("Metadata.channel_id is 'C099VUEKVBN'", metadata["channel_id"], "C099VUEKVBN")
		xtptest.AssertEq("Metadata.channel_name is 'general'", metadata["channel_name"], "general")
		xtptest.AssertEq("Metadata.user is 'sd099rsefgdb_user'", metadata["user"], "sd099rsefgdb_user")
		xtptest.AssertEq("Metadata.thread_ts falls back to message ts", metadata["thread_ts"], "1765611321.248519")
		xtptest.AssertEq("Metadata.team is 'T099VUE950C'", metadata["team"], "T099VUE950C")

		// Verify contexts
		contextsInterface, ok := activity["contexts"]
		if !ok {
			xtptest.AssertEq("Contexts array exists", ok, true)
			return
		}

		contexts, ok := contextsInterface.([]any)
		if !ok {
			xtptest.AssertEq("Contexts is an array", ok, true)
			return
		}

		xtptest.AssertEq("Has 3 contexts", len(contexts), 3)

		// Verify Level 1 - Source context
		source, _ := contexts[0].(map[string]any)
		xtptest.AssertEq("Context[0] ID is 'slack:source'", source["id"], "slack:source")
		xtptest.AssertEq("Context[0] level is 1", int(source["level"].(float64)), 1)
		xtptest.AssertEq("Context[0] resourceType is 'source'", source["resourceType"], "source")

		// Verify Level 2 - Channel context
		channel, _ := contexts[1].(map[string]any)
		xtptest.AssertEq("Context[1] ID is 'slack:channel:C099VUEKVBN'", channel["id"], "slack:channel:C099VUEKVBN")
		xtptest.AssertEq("Context[1] level is 2", int(channel["level"].(float64)), 2)
		xtptest.AssertEq("Context[1] parentId is 'slack:source'", channel["parentId"], "slack:source")
		xtptest.AssertEq("Context[1] resourceType is 'channel'", channel["resourceType"], "channel")

		// Verify channel context has enrichment_params
		channelMetadata, _ := channel["metadata"].(map[string]any)
		enrichmentParams, _ := channelMetadata["enrichment_params"].(map[string]any)
		xtptest.AssertEq("Thread enrichment_params.channel_id", enrichmentParams["channel_id"], "C099VUEKVBN")

		// Verify Level 3 - Thread context (using fallback ts)
		thread, _ := contexts[2].(map[string]any)
		xtptest.AssertEq("Context[2] ID uses fallback ts", thread["id"], "slack:thread:C099VUEKVBN:1765611321.248519")
		xtptest.AssertEq("Context[2] level is 3", int(thread["level"].(float64)), 3)
		xtptest.AssertEq("Context[2] parentId is 'slack:channel:C099VUEKVBN'", thread["parentId"], "slack:channel:C099VUEKVBN")
		xtptest.AssertEq("Context[2] resourceType is 'thread'", thread["resourceType"], "thread")

		// Verify thread context has enrichment_params with fallback ts
		threadMetadata, _ := thread["metadata"].(map[string]any)
		enrichmentParams, _ = threadMetadata["enrichment_params"].(map[string]any)
		xtptest.AssertEq("Thread enrichment_params.channel_id", enrichmentParams["channel_id"], "C099VUEKVBN")
		xtptest.AssertEq("Thread enrichment_params.thread_ts uses fallback", enrichmentParams["thread_ts"], "1765611321.248519")
	})

	return 0
}

func main() {}
