package fetch

import (
	"fmt"
	"github-connector/internal/core"
	"time"
)

type Activity struct {
	ActivityType string
	Contexts     []*core.Context
	Description  *string
	Id           string
	Metadata     any
	Source       string
	Timestamp    time.Time
	Title        string
	Url          *string
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	ref, _ := payload["ref"].(string)
	head, _ := payload["head"].(string)

	title := fmt.Sprintf("Push to %s", repoName)
	description := fmt.Sprintf("Pushed to %s in %s", ref, repoName)
	url := fmt.Sprintf("https://github.com/%s/commit/%s", repoName, head)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

	// Metadata
	metadata := map[string]any{
		"branch":        ref,
		"before_commit": payload["before"],
	}

	// Use ContextGenerator to create hierarchical contexts
	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),             // Level 1: source:github
		gen.CreateRepositoryContext(repoName), // Level 2: repository:{repo}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(payload["number"].(float64))
	action, _ := payload["action"].(string)

	title := fmt.Sprintf("PR #%d %s in %s", prNumber, action, repoName)
	description := fmt.Sprintf("Pull request #%d was %s", prNumber, action)
	url := fmt.Sprintf("https://github.com/%s/pull/%d", repoName, prNumber)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

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

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	issueNumber := int(issue["number"].(float64))
	action, _ := payload["action"].(string)

	title := fmt.Sprintf("Issue #%d %s in %s", issueNumber, action, repoName)
	description := fmt.Sprintf("Issue #%d was %s", issueNumber, action)
	url, _ := issue["html_url"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

	user, _ := issue["user"].(map[string]any)
	labels := extractLabels(issue["labels"])

	metadata := map[string]any{
		"issue_number": issueNumber,
		"action":       action,
		"state":        issue["state"],
		"author":       user["login"],
		"labels":       labels,
	}

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),                     // Level 1: source:github
		gen.CreateRepositoryContext(repoName),         // Level 2: repository:{repo}
		gen.CreateIssueContext(repoName, issueNumber), // Level 3: issue:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(issue["number"].(float64))

	title := fmt.Sprintf("Commented on PR #%d", prNumber)
	description, _ := comment["body"].(string)
	url, _ := comment["html_url"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

	commentUser, _ := comment["user"].(map[string]any)
	metadata := map[string]any{
		"comment_id":         comment["id"],
		"issue_number":       prNumber,
		"comment_author":     commentUser["login"],
		"comment_created_at": comment["created_at"],
	}

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	issueNumber := int(issue["number"].(float64))

	title := fmt.Sprintf("Commented on Issue #%d", issueNumber)
	description, _ := comment["body"].(string)
	url, _ := comment["html_url"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

	commentUser, _ := comment["user"].(map[string]any)
	metadata := map[string]any{
		"comment_id":         comment["id"],
		"issue_number":       issueNumber,
		"comment_author":     commentUser["login"],
		"comment_created_at": comment["created_at"],
	}

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),                     // Level 1: source:github
		gen.CreateRepositoryContext(repoName),         // Level 2: repository:{repo}
		gen.CreateIssueContext(repoName, issueNumber), // Level 3: issue:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	refType, _ := payload["ref_type"].(string)
	ref, _ := payload["ref"].(string)

	title := fmt.Sprintf("Deleted %s %s in %s", refType, ref, repoName)
	description := fmt.Sprintf("%s %s was deleted", refType, ref)
	url := fmt.Sprintf("https://github.com/%s", repoName)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

	metadata := map[string]any{
		"ref_type":    refType,
		"ref":         ref,
		"deleted_by":  actor["login"],
		"pusher_type": payload["pusher_type"],
	}

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),             // Level 1: source:github
		gen.CreateRepositoryContext(repoName), // Level 2: repository:{repo}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(pr["number"].(float64))

	title := fmt.Sprintf("Commented on PR #%d in %s", prNumber, repoName)
	description, _ := comment["body"].(string)
	url, _ := comment["html_url"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

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

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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

	id := core.MakeActivityID(fmt.Sprintf("%v", event["id"]))
	timestampStr, _ := event["created_at"].(string)
	repoName, _ := repo["name"].(string)
	prNumber := int(pr["number"].(float64))

	title := fmt.Sprintf("Reviewed PR #%d in %s", prNumber, repoName)
	description, _ := review["body"].(string)
	url, _ := review["html_url"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}
	timestamp = timestamp.UTC()

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

	gen := core.NewContextGenerator()
	contexts := []*core.Context{
		gen.CreateSourceContext(),               // Level 1: source:github
		gen.CreateRepositoryContext(repoName),   // Level 2: repository:{repo}
		gen.CreatePRContext(repoName, prNumber), // Level 3: pull_request:{number}
	}

	return &Activity{
		Id:           id,
		Timestamp:    timestamp,
		Title:        title,
		Description:  &description,
		Source:       core.ConnectorID,
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
