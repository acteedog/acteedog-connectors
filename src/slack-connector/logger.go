package main

import (
	"fmt"
	"slack-connector/internal/core"

	"github.com/extism/go-pdk"
)

var logger = &Logger{connectorID: core.ConnectorID}

type Logger struct {
	connectorID string
}

func (l *Logger) Error(message string) {
	pdk.Log(pdk.LogError, fmt.Sprintf("[%s] %s", l.connectorID, message))
}

func (l *Logger) Warn(message string) {
	pdk.Log(pdk.LogWarn, fmt.Sprintf("[%s] %s", l.connectorID, message))
}

func (l *Logger) Info(message string) {
	pdk.Log(pdk.LogInfo, fmt.Sprintf("[%s] %s", l.connectorID, message))
}

func (l *Logger) Debug(message string) {
	pdk.Log(pdk.LogDebug, fmt.Sprintf("[%s] %s", l.connectorID, message))
}
