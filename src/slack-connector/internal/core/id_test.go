package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeActivityID(t *testing.T) {
	got := MakeActivityID("1234567890.123456")
	assert.Equal(t, "slack:1234567890.123456", got)
}

func TestMakeSourceContextID(t *testing.T) {
	got := MakeSourceContextID()
	assert.Equal(t, "slack:source", got)
}

func TestMakeChannelContextID(t *testing.T) {
	got := MakeChannelContextID("C1234567890")
	assert.Equal(t, "slack:channel:C1234567890", got)
}

func TestMakeThreadContextID(t *testing.T) {
	got := MakeThreadContextID("C1234567890", "1234567890.123456")
	assert.Equal(t, "slack:thread:C1234567890:1234567890.123456", got)
}
