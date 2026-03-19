package core

// Logger defines the logging interface for connector packages
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// NoopLogger is a logger that discards all messages (used in tests)
type NoopLogger struct{}

func (n *NoopLogger) Info(msg string)  {}
func (n *NoopLogger) Warn(msg string)  {}
func (n *NoopLogger) Error(msg string) {}
