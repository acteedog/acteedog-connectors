package main

import (
	"encoding/json"

	xtptest "github.com/dylibso/xtp-test-go"
)

//go:export test
func test() int32 {
	xtptest.Group("ApplyRepositoryEnrichment", func() {
		apiResponseJSON := xtptest.MockInputBytes()

		outputJSON := xtptest.CallBytes("ApplyRepositoryEnrichment", apiResponseJSON)

		var context map[string]any
		err := json.Unmarshal(outputJSON, &context)
		xtptest.Assert("Context can be parsed as JSON", err == nil, "Failed to parse context JSON: "+string(outputJSON))
		if err != nil {
			return
		}

		// Verify context
		xtptest.AssertEq("ConnectorId is 'github'", context["connectorId"], "github")
		xtptest.AssertEq("Description is enriched", context["description"], "My awesome test repo")
		xtptest.AssertEq("Id is correct", context["id"], "github:repository:testorg/testrepo")

		level, _ := context["level"].(float64)
		xtptest.AssertEq("Level is 2", int(level), 2)

		xtptest.AssertEq("Name is correct", context["name"], "repository:testorg/testrepo")
		xtptest.AssertEq("ParentId is correct", context["parentId"], "github:source")
		xtptest.AssertEq("ResourceType is 'repository'", context["resourceType"], "repository")
		xtptest.AssertEq("Title is enriched", context["title"], "Repository: testorg/testrepo")
		xtptest.AssertEq("Url is enriched", context["url"], "https://github.com/testorg/testrepo")
		xtptest.AssertEq("createdAt is enriched", context["createdAt"], "2015-02-13T07:54:25Z")
		xtptest.AssertEq("updatedAt is enriched", context["updatedAt"], "2025-12-05T10:30:01Z")

		// Verify metadata
		metadata, ok := context["metadata"].(map[string]any)
		xtptest.Assert("metadata is a map", ok, "metadata is not a map")
		if ok {
			stargazers, _ := metadata["stargazers_count"].(float64)
			xtptest.AssertEq("stargazers_count is 13", int(stargazers), 13)

			xtptest.AssertEq("language is 'Go'", metadata["language"], "Go")

			topics, ok := metadata["topics"].([]any)
			xtptest.Assert("topics is an array", ok, "topics is not an array")
			if ok {
				xtptest.AssertEq("topics has 1 item", len(topics), 1)
				xtptest.AssertEq("topic is 'ruby-on-rails'", topics[0], "ruby-on-rails")
			}

			xtptest.AssertEq("default_branch is 'main'", metadata["default_branch"], "main")
			xtptest.AssertEq("visibility is 'public'", metadata["visibility"], "public")

			forks, _ := metadata["forks_count"].(float64)
			xtptest.AssertEq("forks_count is 216", int(forks), 216)

			issues, _ := metadata["open_issues_count"].(float64)
			xtptest.AssertEq("open_issues_count is 216", int(issues), 216)

			watchers, _ := metadata["watchers_count"].(float64)
			xtptest.AssertEq("watchers_count is 13", int(watchers), 13)

			xtptest.AssertEq("homepage is correct", metadata["homepage"], "https://example.com")

			// Verify enrichment_params for repository
			params, ok := metadata["enrichment_params"].(map[string]any)
			xtptest.Assert("metadata has enrichment_params", ok, "enrichment_params not found")
			if ok {
				repo, _ := params["repo"]
				xtptest.AssertEq("repo is correct", repo, "testorg/testrepo")
			}
		}
	})

	return 0
}

func main() {}
