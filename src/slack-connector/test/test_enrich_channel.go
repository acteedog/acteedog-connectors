package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("ApplyChannelEnrichment", func() {
		apiResponseJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("ApplyChannelEnrichment", apiResponseJSON)

		var context map[string]any
		if err := json.Unmarshal(outputJSON, &context); err != nil {
			xtptest.AssertEq("Context can be parsed as JSON", err, nil)
			return
		}

		// Verify context
		xtptest.AssertEq("ConnectorId is 'slack'", context["connectorId"], "slack")
		xtptest.AssertEq("Description contains topic", context["description"], "Company-wide announcements and work-based matters")
		xtptest.AssertEq("Id is 'slack:channel:C099VUEKVBN'", context["id"], "slack:channel:C099VUEKVBN")
		xtptest.AssertEq("Level is 2", int(context["level"].(float64)), 2)
		xtptest.AssertEq("Name is '#general'", context["name"], "channel #general")
		xtptest.AssertEq("ParentId is 'slack:source'", context["parentId"], "slack:source")
		xtptest.AssertEq("ResourceType is 'channel'", context["resourceType"], "channel")
		xtptest.AssertEq("Title is '#general'", context["title"], "#general")
		xtptest.AssertEq("URL is correctly formatted", context["url"], "https://test-workspace.slack.com/archives/C099VUEKVBN")
		xtptest.AssertEq("CreatedAt is correct", context["createdAt"], "2025-08-11T02:02:56Z")
		xtptest.AssertEq("UpdatedAt is correct", context["updatedAt"], "2025-12-08T14:33:49.239Z")

		// Verify metadata
		metadata, ok := context["metadata"].(map[string]any)
		if !ok {
			xtptest.AssertEq("Metadata exists and is a map", ok, true)
			return
		}

		xtptest.AssertEq("Metadata.name is 'general'", metadata["name"], "general")
		xtptest.AssertEq("Metadata.is_private is false", metadata["is_private"], false)
		xtptest.AssertEq("Metadata.is_channel is true", metadata["is_channel"], true)
		xtptest.AssertEq("Metadata.is_group is false", metadata["is_group"], false)
		xtptest.AssertEq("Metadata.is_im is false", metadata["is_im"], false)
		xtptest.AssertEq("Metadata.topic exists", metadata["topic"], "Company-wide announcements and work-based matters")
		xtptest.AssertEq("Metadata.purpose exists", metadata["purpose"], "This channel is for workspace-wide communication and announcements. All members are in this channel.")
		xtptest.AssertEq("Metadata.context_team_id is 'T099VUE950C'", metadata["context_team_id"], "T099VUE950C")

		created, _ := metadata["created"].(float64)
		xtptest.AssertEq("Metadata.created is 1754877776", int64(created), int64(1754877776))

		// Verify enrichment_params
		enrichmentParams, ok := metadata["enrichment_params"].(map[string]any)
		xtptest.AssertEq("Metadata.enrichment_params exists and is a map", ok, true)
		if !ok {
			return
		}
		xtptest.AssertEq("enrichment_params.channel_id is 'C099VUEKVBN'", enrichmentParams["channel_id"], "C099VUEKVBN")
	})

	return 0
}

func main() {}
