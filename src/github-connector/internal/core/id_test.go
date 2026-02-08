package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeActivityID(t *testing.T) {
	assert.Equal(t, "github:evt_12345", MakeActivityID("evt_12345"))
}

func TestMakeSourceContextID(t *testing.T) {
	assert.Equal(t, "github:source", MakeSourceContextID())
}

func TestMakeRepositoryContextID(t *testing.T) {
	assert.Equal(t, "github:repository:owner/repo", MakeRepositoryContextID("owner/repo"))
}

func TestMakePullRequestContextID(t *testing.T) {
	assert.Equal(t, "github:pull_request:owner/repo:42", MakePullRequestContextID("owner/repo", "42"))
}

func TestMakeIssueContextID(t *testing.T) {
	assert.Equal(t, "github:issue:owner/repo:101", MakeIssueContextID("owner/repo", "101"))
}
