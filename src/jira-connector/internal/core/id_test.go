package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeIssueCreatedActivityID(t *testing.T) {
	assert.Equal(t, "jira:project:10000:issue:10038:created", MakeIssueCreatedActivityID("10000", "10038"))
}

func TestMakeCommentActivityID(t *testing.T) {
	assert.Equal(t, "jira:project:10000:issue:10003:comment:10000", MakeCommentActivityID("10000", "10003", "10000"))
}

func TestMakeStatusChangedActivityID(t *testing.T) {
	assert.Equal(t, "jira:project:10000:issue:10003:status_changed:10045", MakeStatusChangedActivityID("10000", "10003", "10045"))
}

func TestMakeSourceContextID(t *testing.T) {
	assert.Equal(t, "jira:source", MakeSourceContextID())
}

func TestMakeProjectContextID(t *testing.T) {
	assert.Equal(t, "jira:project:10000", MakeProjectContextID("10000"))
}

func TestMakeIssueContextID(t *testing.T) {
	assert.Equal(t, "jira:project:10000:issue:10038", MakeIssueContextID("10000", "10038"))
}
