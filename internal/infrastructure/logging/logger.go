package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger with convenience methods.
type Logger struct {
	*slog.Logger
}

// LogLevel represents logging levels.
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// New creates a new structured logger with JSON output to stderr.
func New(level LogLevel) *Logger {
	var slogLevel slog.Level
	switch strings.ToLower(string(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewJSONHandler(os.Stderr, opts)
	return &Logger{slog.New(handler)}
}

// NewNoop creates a logger that discards all output.
func NewNoop() *Logger {
	return &Logger{slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.Level(100), // Higher than any level to discard all
	}))}
}

// With returns a new Logger with the given attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

// WithComponent returns a logger with a component attribute.
func (l *Logger) WithComponent(name string) *Logger {
	return l.With("component", name)
}

// ParseLevel parses a string into a LogLevel.
func ParseLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
