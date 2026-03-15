package core

// Logger defines a simple logging interface
type Logger interface {
	Error(string)
	Warn(string)
	Info(string)
	Debug(string)
}

// NoopLogger is a logger that does nothing
type NoopLogger struct{}

// NewNoopLogger creates a new NoopLogger instance
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// Error logs an error message (no-op)
func (l *NoopLogger) Error(msg string) {}

// Warn logs a warning message (no-op)
func (l *NoopLogger) Warn(msg string) {}

// Info logs an info message (no-op)
func (l *NoopLogger) Info(msg string) {}

// Debug logs a debug message (no-op)
func (l *NoopLogger) Debug(msg string) {}
