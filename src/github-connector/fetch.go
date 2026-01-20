package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/extism/go-pdk"
)

// FetchActivities fetches GitHub activities based on the input configuration and parameters
func FetchActivities(input FetchRequest) (FetchResponse, error) {
	pdk.Log(pdk.LogInfo, "FetchActivities: Starting GitHub events fetch")

	config, ok := input.Config.(map[string]any)
	if !ok {
		return FetchResponse{}, fmt.Errorf("invalid configuration format")
	}

	token, ok := config["credential_personal_access_token"].(string)
	if !ok || token == "" {
		return FetchResponse{}, fmt.Errorf("missing personal access token")
	}

	username, ok := config["username"].(string)
	if !ok || username == "" {
		return FetchResponse{}, fmt.Errorf("missing username")
	}

	var repositoryPatterns []string
	if patternsInterface, ok := config["repository_patterns"]; ok && patternsInterface != nil {
		if patterns, ok := patternsInterface.([]any); ok {
			for _, p := range patterns {
				if patternStr, ok := p.(string); ok && patternStr != "" {
					repositoryPatterns = append(repositoryPatterns, patternStr)
				}
			}
		}
	}

	if len(repositoryPatterns) > 0 {
		pdk.Log(pdk.LogInfo, fmt.Sprintf("Using %d repository patterns", len(repositoryPatterns)))
	} else {
		pdk.Log(pdk.LogInfo, "No repository patterns specified, fetching all repositories")
	}

	targetDate := input.Params.TargetDate
	pdk.Log(pdk.LogInfo, fmt.Sprintf("Fetching events for user: %s, date: %s", username, targetDate))

	startTime, endTime, err := parseDateRange(targetDate)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("invalid target date: %w", err)
	}

	allEvents, err := fetchAllEvents(username, token, startTime, endTime)
	if err != nil {
		return FetchResponse{}, err
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Fetched %d events", len(allEvents)))

	filteredEvents := filterEventsByRepository(allEvents, repositoryPatterns)
	pdk.Log(pdk.LogInfo, fmt.Sprintf("After repository filtering: %d events", len(filteredEvents)))

	activities := []Activity{}
	for _, event := range filteredEvents {
		activity, err := transformEvent(event)
		if err != nil {
			pdk.Log(pdk.LogDebug, fmt.Sprintf("Skipping event: %s", err.Error()))
			continue
		}
		if activity != nil {
			activities = append(activities, *activity)
		}
	}

	pdk.Log(pdk.LogInfo, fmt.Sprintf("Transformed %d activities", len(activities)))

	return FetchResponse{
		Activities: activities,
	}, nil
}

// parseDateRange parses the target date and returns start and end times for the day
func parseDateRange(targetDate string) (time.Time, time.Time, error) {
	var t time.Time
	var err error

	// Try RFC3339 format first (2025-12-12T00:00:00Z or 2025-12-12T00:00:00+00:00)
	t, err = time.Parse(time.RFC3339, targetDate)
	if err != nil {
		// Fallback to date-only format (2025-12-12)
		t, err = time.Parse("2006-01-02", targetDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	endTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)

	return startTime, endTime, nil
}

// fetchAllEvents fetches events from GitHub API with pagination
func fetchAllEvents(username, token string, startTime, endTime time.Time) ([]map[string]any, error) {
	allEvents := []map[string]any{}

	// GitHub Events API returns max 300 events (3 pages with per_page=100)
	for page := 1; page <= 3; page++ {
		url := fmt.Sprintf("https://api.github.com/users/%s/events?per_page=100&page=%d", username, page)

		pdk.Log(pdk.LogDebug, fmt.Sprintf("Fetching page %d: %s", page, url))

		req := pdk.NewHTTPRequest(pdk.MethodGet, url)
		req.SetHeader("Authorization", "token "+token)
		req.SetHeader("Accept", "application/vnd.github+json")
		req.SetHeader("User-Agent", "acteedog/"+ConnectorID)

		res := req.Send()

		if res.Status() != 200 {
			return nil, fmt.Errorf("GitHub API error: HTTP %d", res.Status())
		}

		var events []map[string]any
		if err := json.Unmarshal(res.Body(), &events); err != nil {
			return nil, fmt.Errorf("failed to parse events: %w", err)
		}

		if len(events) == 0 {
			pdk.Log(pdk.LogDebug, "No more events, stopping pagination")
			break
		}

		// Filter events by date and check if we should stop
		filteredEvents, shouldStop := filterEventsByDate(events, startTime, endTime)
		allEvents = append(allEvents, filteredEvents...)

		pdk.Log(pdk.LogDebug, fmt.Sprintf("Page %d: %d events fetched, %d filtered", page, len(events), len(filteredEvents)))

		if shouldStop {
			pdk.Log(pdk.LogDebug, "Reached events outside date range, stopping pagination")
			break
		}
	}

	return allEvents, nil
}

// filterEventsByDate filters events by the target date range
func filterEventsByDate(events []map[string]any, startTime, endTime time.Time) ([]map[string]any, bool) {
	filtered := []map[string]any{}
	shouldStop := false

	for _, event := range events {
		createdAtStr, ok := event["created_at"].(string)
		if !ok {
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			continue
		}

		if createdAt.After(startTime) && createdAt.Before(endTime) {
			filtered = append(filtered, event)
		}

		if createdAt.Before(startTime) {
			shouldStop = true
			break
		}
	}

	return filtered, shouldStop
}

// filterEventsByRepository filters events by repository patterns
func filterEventsByRepository(events []map[string]any, patterns []string) []map[string]any {
	if len(patterns) == 0 {
		return events
	}

	filtered := []map[string]any{}
	for _, event := range events {
		repo, ok := event["repo"].(map[string]any)
		if !ok {
			continue
		}

		repoName, ok := repo["name"].(string)
		if !ok {
			continue
		}

		if matchesAnyPattern(repoName, patterns) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// matchesAnyPattern checks if a repository name matches any of the patterns
func matchesAnyPattern(repoName string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(repoName, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if a string matches a pattern with * wildcards
// Supports patterns like:
// - "owner/*" (matches all repos in owner)
// - "owner/repo" (exact match)
// - "owner/prefix-*" (matches repos starting with prefix-)
// - "*/*" (matches all repos)
func matchPattern(str, pattern string) bool {
	// Convert pattern to regex-like matching
	// Split both by '/' to handle owner/repo structure
	strParts := splitRepo(str)
	patternParts := splitRepo(pattern)

	// Must have same structure (owner/repo)
	if len(strParts) != len(patternParts) {
		return false
	}

	// Check each part
	for i := range strParts {
		if !matchWildcard(strParts[i], patternParts[i]) {
			return false
		}
	}

	return true
}

// splitRepo splits repository name by '/'
func splitRepo(repo string) []string {
	parts := []string{}
	current := ""
	for _, ch := range repo {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// matchWildcard checks if a string matches a pattern with * wildcard
// * matches one or more characters (not empty string)
func matchWildcard(str, pattern string) bool {
	// Exact match
	if pattern == str {
		return true
	}

	// No wildcard
	if !containsWildcard(pattern) {
		return false
	}

	// Handle wildcard matching
	if pattern == "*" {
		return len(str) > 0
	}

	// Pattern with wildcard at the end (e.g., "prefix-*")
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(str) >= len(prefix) && str[:len(prefix)] == prefix
	}

	// Pattern with wildcard at the start (e.g., "*-suffix")
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(str) >= len(suffix) && str[len(str)-len(suffix):] == suffix
	}

	// Pattern with wildcard in the middle (e.g., "pre*fix")
	// Find the position of '*'
	starPos := -1
	for i, ch := range pattern {
		if ch == '*' {
			starPos = i
			break
		}
	}

	if starPos >= 0 {
		prefix := pattern[:starPos]
		suffix := pattern[starPos+1:]

		if len(str) < len(prefix)+len(suffix) {
			return false
		}

		return str[:len(prefix)] == prefix && str[len(str)-len(suffix):] == suffix
	}

	return false
}

// containsWildcard checks if a string contains '*'
func containsWildcard(s string) bool {
	for _, ch := range s {
		if ch == '*' {
			return true
		}
	}
	return false
}

// transformEvent transforms a GitHub event to an Activity
func transformEvent(event map[string]any) (*Activity, error) {
	eventType, ok := event["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing event type")
	}

	switch eventType {
	case "PushEvent":
		return transformPushEvent(event)
	case "PullRequestEvent":
		return transformPullRequestEvent(event)
	case "IssuesEvent":
		return transformIssuesEvent(event)
	case "IssueCommentEvent":
		return transformIssueCommentEvent(event)
	case "DeleteEvent":
		return transformDeleteEvent(event)
	case "PullRequestReviewCommentEvent":
		return transformPRReviewCommentEvent(event)
	case "PullRequestReviewEvent":
		return transformPRReviewEvent(event)
	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}

// transformPushEvent transforms a PushEvent to an Activity
func transformPushEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in PushEvent")
	}

	repo, ok := event["repo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid repo in PushEvent")
	}

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	ref, _ := payload["ref"].(string)
	head, _ := payload["head"].(string)

	title := fmt.Sprintf("Push to %s", repoName)
	description := fmt.Sprintf("Pushed to %s in %s", ref, repoName)
	url := fmt.Sprintf("https://github.com/%s/commit/%s", repoName, head)

	// Metadata
	metadata := map[string]any{
		"branch":        ref,
		"before_commit": payload["before"],
	}

	// Use ContextGenerator to create hierarchical contexts
	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),             // Level 1: source:github
		gen.CreateRepositoryContext(repoName), // Level 2: repository:{repo}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "push",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformPullRequestEvent transforms a PullRequestEvent to an Activity
func transformPullRequestEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in PullRequestEvent")
	}

	repo, ok := event["repo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid repo in PullRequestEvent")
	}

	pr, ok := payload["pull_request"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid pull_request in PullRequestEvent")
	}

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(payload["number"].(float64))
	action, _ := payload["action"].(string)

	title := fmt.Sprintf("PR #%d %s in %s", prNumber, action, repoName)
	description := fmt.Sprintf("Pull request #%d was %s", prNumber, action)
	url := fmt.Sprintf("https://github.com/%s/pull/%d", repoName, prNumber)

	base, _ := pr["base"].(map[string]any)
	head, _ := pr["head"].(map[string]any)
	metadata := map[string]any{
		"pr_number":   prNumber,
		"action":      action,
		"base_branch": base["ref"],
		"head_branch": head["ref"],
		"base_sha":    base["sha"],
		"head_sha":    head["sha"],
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "pull_request",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformIssuesEvent transforms an IssuesEvent to an Activity
func transformIssuesEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in IssuesEvent")
	}

	repo, ok := event["repo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid repo in IssuesEvent")
	}

	issue, ok := payload["issue"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid issue in IssuesEvent")
	}

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	issueNumber := int(issue["number"].(float64))
	action, _ := payload["action"].(string)

	title := fmt.Sprintf("Issue #%d %s in %s", issueNumber, action, repoName)
	description := fmt.Sprintf("Issue #%d was %s", issueNumber, action)
	url, _ := issue["html_url"].(string)

	user, _ := issue["user"].(map[string]any)
	labels := extractLabels(issue["labels"])

	metadata := map[string]any{
		"issue_number": issueNumber,
		"action":       action,
		"state":        issue["state"],
		"author":       user["login"],
		"labels":       labels,
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),                     // Level 1: source:github
		gen.CreateRepositoryContext(repoName),         // Level 2: repository:{repo}
		gen.CreateIssueContext(repoName, issueNumber), // Level 3: issue:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "issues",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformIssueCommentEvent transforms an IssueCommentEvent to an Activity
// Distinguishes between PR comments and issue comments
func transformIssueCommentEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in IssueCommentEvent")
	}

	issue, ok := payload["issue"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid issue in IssueCommentEvent")
	}

	// Check if this is a PR comment or issue comment
	_, isPR := issue["pull_request"]

	if isPR {
		return transformPRCommentEvent(event)
	}
	return transformIssueCommentOnlyEvent(event)
}

// transformPRCommentEvent transforms a PR comment (IssueCommentEvent on PR)
func transformPRCommentEvent(event map[string]any) (*Activity, error) {
	payload, _ := event["payload"].(map[string]any)
	repo, _ := event["repo"].(map[string]any)
	issue, _ := payload["issue"].(map[string]any)
	comment, _ := payload["comment"].(map[string]any)

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(issue["number"].(float64))

	title := fmt.Sprintf("Commented on PR #%d", prNumber)
	description, _ := comment["body"].(string)
	url, _ := comment["html_url"].(string)

	commentUser, _ := comment["user"].(map[string]any)
	metadata := map[string]any{
		"comment_id":         comment["id"],
		"issue_number":       prNumber,
		"comment_author":     commentUser["login"],
		"comment_created_at": comment["created_at"],
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "pr_comment",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformIssueCommentOnlyEvent transforms an issue comment (IssueCommentEvent on Issue)
func transformIssueCommentOnlyEvent(event map[string]any) (*Activity, error) {
	payload, _ := event["payload"].(map[string]any)
	repo, _ := event["repo"].(map[string]any)
	issue, _ := payload["issue"].(map[string]any)
	comment, _ := payload["comment"].(map[string]any)

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	issueNumber := int(issue["number"].(float64))

	title := fmt.Sprintf("Commented on Issue #%d", issueNumber)
	description, _ := comment["body"].(string)
	url, _ := comment["html_url"].(string)

	commentUser, _ := comment["user"].(map[string]any)
	metadata := map[string]any{
		"comment_id":         comment["id"],
		"issue_number":       issueNumber,
		"comment_author":     commentUser["login"],
		"comment_created_at": comment["created_at"],
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),                     // Level 1: source:github
		gen.CreateRepositoryContext(repoName),         // Level 2: repository:{repo}
		gen.CreateIssueContext(repoName, issueNumber), // Level 3: issue:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "issue_comment",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformDeleteEvent transforms a DeleteEvent to an Activity
func transformDeleteEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in DeleteEvent")
	}

	repo, ok := event["repo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid repo in DeleteEvent")
	}

	actor, _ := event["actor"].(map[string]any)

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	refType, _ := payload["ref_type"].(string)
	ref, _ := payload["ref"].(string)

	title := fmt.Sprintf("Deleted %s %s in %s", refType, ref, repoName)
	description := fmt.Sprintf("%s %s was deleted", refType, ref)
	url := fmt.Sprintf("https://github.com/%s", repoName)

	metadata := map[string]any{
		"ref_type":    refType,
		"ref":         ref,
		"deleted_by":  actor["login"],
		"pusher_type": payload["pusher_type"],
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),             // Level 1: source:github
		gen.CreateRepositoryContext(repoName), // Level 2: repository:{repo}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "delete",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformPRReviewCommentEvent transforms a PullRequestReviewCommentEvent to an Activity
func transformPRReviewCommentEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in PullRequestReviewCommentEvent")
	}

	repo, ok := event["repo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid repo in PullRequestReviewCommentEvent")
	}

	pr, ok := payload["pull_request"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid pull_request in PullRequestReviewCommentEvent")
	}

	comment, ok := payload["comment"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid comment in PullRequestReviewCommentEvent")
	}

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(pr["number"].(float64))

	title := fmt.Sprintf("Commented on PR #%d in %s", prNumber, repoName)
	description, _ := comment["body"].(string)
	url, _ := comment["html_url"].(string)

	commentUser, _ := comment["user"].(map[string]any)
	prBase, _ := pr["base"].(map[string]any)
	prHead, _ := pr["head"].(map[string]any)

	metadata := map[string]any{
		"comment_id":     comment["id"],
		"pr_number":      prNumber,
		"comment_author": commentUser["login"],
		"file_path":      comment["path"],
		"commit_id":      comment["commit_id"],
		"base_branch":    prBase["ref"],
		"head_branch":    prHead["ref"],
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "pr_review_comment",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// transformPRReviewEvent transforms a PullRequestReviewEvent to an Activity
func transformPRReviewEvent(event map[string]any) (*Activity, error) {
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payload in PullRequestReviewEvent")
	}

	repo, ok := event["repo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid repo in PullRequestReviewEvent")
	}

	pr, ok := payload["pull_request"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid pull_request in PullRequestReviewEvent")
	}

	review, ok := payload["review"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid review in PullRequestReviewEvent")
	}

	id := makeActivityID(fmt.Sprintf("%v", event["id"]))
	timestamp, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(pr["number"].(float64))

	title := fmt.Sprintf("Reviewed PR #%d in %s", prNumber, repoName)
	description, _ := review["body"].(string)
	url, _ := review["html_url"].(string)

	reviewer, _ := review["user"].(map[string]any)
	prBase, _ := pr["base"].(map[string]any)
	prHead, _ := pr["head"].(map[string]any)

	metadata := map[string]any{
		"pr_number":    prNumber,
		"review_state": review["state"],
		"reviewer":     reviewer["login"],
		"submitted_at": review["submitted_at"],
		"base_branch":  prBase["ref"],
		"head_branch":  prHead["ref"],
	}

	gen := NewContextGenerator()
	contexts := []Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       ConnectorID,
		ActivityType: "pr_review",
		Url:          &url,
		Metadata:     metadata,
		Contexts:     contexts,
	}, nil
}

// extractLabels extracts label names from issue labels array
func extractLabels(labelsInterface any) []string {
	labels := []string{}
	if labelsList, ok := labelsInterface.([]any); ok {
		for _, labelItem := range labelsList {
			if label, ok := labelItem.(map[string]any); ok {
				if name, ok := label["name"].(string); ok {
					labels = append(labels, name)
				}
			}
		}
	}
	return labels
}

// ============================================================================
// Test Exports
// ============================================================================

//go:export TransformEvent
func TransformEvent() int32 {
	input := pdk.Input()

	var event map[string]any
	if err := json.Unmarshal(input, &event); err != nil {
		pdk.SetError(err)
		return 1
	}

	activity, err := transformEvent(event)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	output, err := json.Marshal(activity)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.Output(output)
	return 0
}
