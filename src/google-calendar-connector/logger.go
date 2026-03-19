//go:build wasip1

package main

import (
	"google-calendar-connector/internal/core"

	"github.com/extism/go-pdk"
)

// Logger wraps pdk.Log to implement core.Logger
type Logger struct {
	connectorID string
}

func (l *Logger) Info(msg string) {
	pdk.Log(pdk.LogInfo, msg)
}

func (l *Logger) Warn(msg string) {
	pdk.Log(pdk.LogWarn, msg)
}

func (l *Logger) Error(msg string) {
	pdk.Log(pdk.LogError, msg)
}

var logger core.Logger = &Logger{connectorID: core.ConnectorID}
