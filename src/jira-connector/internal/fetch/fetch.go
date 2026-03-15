package fetch

import (
	"fmt"
	"jira-connector/internal/core"
	"time"
)

// Activity represents an activity fetched from Jira
type Activity struct {
	ActivityType string
	Contexts     []*core.Context
	Description  string
	Id           string
	Metadata     any
	Source       string
	Timestamp    time.Time
	Title        string
	Url          *string
}

// ActivityFetcher defines the structure for fetching activities from Jira
type ActivityFetcher struct {
	httpClient HTTPClient
	config     *config
	logger     core.Logger
}

// NewActivityFetcher creates a new ActivityFetcher instance
func NewActivityFetcher(httpClient HTTPClient, cfg any, targetDate string, logger core.Logger) (*ActivityFetcher, error) {
	config, err := newConfig(cfg, targetDate)
	if err != nil {
		return nil, err
	}

	return &ActivityFetcher{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}, nil
}

// FetchActivities fetches and processes activities from Jira
func (f *ActivityFetcher) FetchActivities() ([]*Activity, error) {
	f.logger.Info("Starting to fetch Jira issues")

	nextDate := f.config.startTime.AddDate(0, 0, 1).Format("2006-01-02")
	response, err := f.httpClient.FetchIssues(
		f.config.CloudID,
		f.config.Email,
		f.config.APIToken,
		f.config.ProjectIDs,
		f.config.targetDate,
		nextDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	if response == nil || len(response.Issues) == 0 {
		f.logger.Info("No issues found in response")
		return []*Activity{}, nil
	}

	f.logger.Info(fmt.Sprintf("Fetched %d issues", len(response.Issues)))

	cgen := core.NewContextGenerator(f.config.CloudID)
	activities := []*Activity{}

	for i := range response.Issues {
		issueActivities, err := f.transformIssue(&response.Issues[i], cgen)
		if err != nil {
			f.logger.Warn(fmt.Sprintf("Skipping issue: %s", err.Error()))
			continue
		}
		activities = append(activities, issueActivities...)
	}

	f.logger.Info(fmt.Sprintf("Transformed %d activities", len(activities)))

	return activities, nil
}

// transformIssue converts a Jira issue into a list of activities
func (f *ActivityFetcher) transformIssue(issue *JiraIssue, cgen *core.ContextGenerator) ([]*Activity, error) {
	if issue.ID == "" || issue.Key == "" {
		return nil, fmt.Errorf("issue missing id or key")
	}

	issueID := issue.ID
	issueKey := issue.Key
	summary := issue.Fields.Summary

	sourceCtx := cgen.CreateSourceContext()
	var projectCtx *core.Context
	var parentCtx *core.Context
	var issueCtx *core.Context

	if issue.Fields.Project == nil || issue.Fields.Project.ID == "" {
		return nil, fmt.Errorf("issue %s missing project field", issueKey)
	}
	projectID := issue.Fields.Project.ID
	projectName := issue.Fields.Project.Name
	projectCtx = cgen.CreateProjectContext(projectID, projectName)

	if issue.Fields.Parent != nil && issue.Fields.Parent.ID != "" {
		parentID := issue.Fields.Parent.ID
		parentKey := issue.Fields.Parent.Key
		parentSummary := ""
		parentIssueTypeName := ""
		if issue.Fields.Parent.Fields != nil {
			parentSummary = issue.Fields.Parent.Fields.Summary
			if issue.Fields.Parent.Fields.IssueType != nil {
				parentIssueTypeName = issue.Fields.Parent.Fields.IssueType.Name
			}
		}
		issueTypeName := ""
		if issue.Fields.IssueType != nil {
			issueTypeName = issue.Fields.IssueType.Name
		}
		parentCtx = cgen.CreateParentIssueContext(parentID, parentKey, parentSummary, projectID, parentIssueTypeName)
		issueCtx = cgen.CreateIssueContextWithIssueParent(issueID, issueKey, summary, projectID, parentID, issueTypeName)
	} else {
		issueTypeName := ""
		if issue.Fields.IssueType != nil {
			issueTypeName = issue.Fields.IssueType.Name
		}
		issueCtx = cgen.CreateIssueContextWithProjectParent(issueID, issueKey, summary, projectID, issueTypeName)
	}

	var contexts []*core.Context
	if parentCtx != nil {
		contexts = []*core.Context{sourceCtx, projectCtx, parentCtx, issueCtx}
	} else {
		contexts = []*core.Context{sourceCtx, projectCtx, issueCtx}
	}

	var activities []*Activity

	// Issue creation activity
	if issue.Fields.Creator != nil && issue.Fields.Creator.EmailAddress == f.config.Email && issue.Fields.Created != "" {
		createdTime, err := core.ParseJiraTime(issue.Fields.Created)
		if err == nil {
			createdTime = createdTime.UTC()
			if isOnTargetDate(createdTime, f.config.startTime, f.config.endTime) {
				title := fmt.Sprintf("Created issue %s: %s", issueKey, summary)
				issueURL := ptrStr(fmt.Sprintf("https://%s.atlassian.net/browse/%s", f.config.SiteSubdomain, issueKey))
				activities = append(activities, &Activity{
					Id:           core.MakeIssueCreatedActivityID(projectID, issueID),
					Timestamp:    createdTime,
					Source:       core.ConnectorID,
					ActivityType: "created",
					Title:        title,
					Url:          issueURL,
					Metadata: map[string]any{
						"issue_id":  issueID,
						"issue_key": issueKey,
					},
					Contexts: contexts,
				})
			}
		}
	}

	// Comment activities
	if issue.Fields.Comment != nil {
		for i := range issue.Fields.Comment.Comments {
			comment := &issue.Fields.Comment.Comments[i]
			if comment.ID == "" || comment.Author == nil || comment.Created == "" {
				continue
			}
			if comment.Author.EmailAddress != f.config.Email {
				continue
			}

			commentCreatedTime, err := core.ParseJiraTime(comment.Created)
			if err != nil {
				continue
			}
			commentCreatedTime = commentCreatedTime.UTC()

			if !isOnTargetDate(commentCreatedTime, f.config.startTime, f.config.endTime) {
				continue
			}

			commentBody := comment.Body.PlainText()
			title := fmt.Sprintf("Commented on %s: %s", issueKey, summary)
			commentURL := ptrStr(fmt.Sprintf("https://%s.atlassian.net/browse/%s?focusedCommentId=%s", f.config.SiteSubdomain, issueKey, comment.ID))
			activities = append(activities, &Activity{
				Id:           core.MakeCommentActivityID(projectID, issueID, comment.ID),
				Timestamp:    commentCreatedTime,
				Source:       core.ConnectorID,
				ActivityType: "commented",
				Title:        title,
				Description:  commentBody,
				Url:          commentURL,
				Metadata: map[string]any{
					"issue_id":   issueID,
					"issue_key":  issueKey,
					"comment_id": comment.ID,
				},
				Contexts: contexts,
			})
		}
	}

	// Status change activities
	for i := range issue.Changelog.Histories {
		history := &issue.Changelog.Histories[i]
		if history.ID == "" || history.Author == nil || history.Created == "" {
			continue
		}
		if history.Author.EmailAddress != f.config.Email {
			continue
		}

		historyCreatedTime, err := core.ParseJiraTime(history.Created)
		if err != nil {
			continue
		}
		historyCreatedTime = historyCreatedTime.UTC()

		if !isOnTargetDate(historyCreatedTime, f.config.startTime, f.config.endTime) {
			continue
		}

		// Only include status changes
		for j := range history.Items {
			item := &history.Items[j]
			if item.Field != "status" {
				continue
			}

			fromStatus := item.FromString
			toStatus := item.ToString
			title := fmt.Sprintf("Changed status of %s from %s to %s", issueKey, fromStatus, toStatus)
			issueURL := ptrStr(fmt.Sprintf("https://%s.atlassian.net/browse/%s", f.config.SiteSubdomain, issueKey))
			activities = append(activities, &Activity{
				Id:           core.MakeStatusChangedActivityID(projectID, issueID, history.ID),
				Timestamp:    historyCreatedTime,
				Source:       core.ConnectorID,
				ActivityType: "status_changed",
				Title:        title,
				Url:          issueURL,
				Metadata: map[string]any{
					"issue_id":    issueID,
					"issue_key":   issueKey,
					"from_status": fromStatus,
					"to_status":   toStatus,
				},
				Contexts: contexts,
			})
			break // Only one status change per history entry
		}
	}

	return activities, nil
}

// isOnTargetDate checks if a time falls within the target date range
func isOnTargetDate(t, startTime, endTime time.Time) bool {
	return !t.Before(startTime) && !t.After(endTime)
}

// ptrStr returns a pointer to the given string
func ptrStr(s string) *string {
	return &s
}
