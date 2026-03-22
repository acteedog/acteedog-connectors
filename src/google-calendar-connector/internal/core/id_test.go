package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeActivityID(t *testing.T) {
	assert.Equal(t, "google-calendar:primary:abc123", MakeActivityID("primary", "abc123"))
}

func TestMakeSourceContextID(t *testing.T) {
	assert.Equal(t, "google-calendar:source", MakeSourceContextID())
}

func TestMakeCalendarContextID(t *testing.T) {
	assert.Equal(t, "google-calendar:calendar:primary", MakeCalendarContextID("primary"))
}

func TestMakeEventContextID(t *testing.T) {
	assert.Equal(t, "google-calendar:event:primary:abc123", MakeEventContextID("primary", "abc123"))
}
