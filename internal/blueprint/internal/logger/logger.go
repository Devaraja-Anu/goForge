package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

type contextKey struct{}

var loggerKey = &contextKey{}

func New() *slog.Logger {
	return NewWithWriter(os.Stdout)
}

// NewWithWriter builds a logger writing to w instead of stdout. Primarily for tests.
func NewWithWriter(w io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("APP_ENV") == "development" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewJSONHandler(w, opts)

	return slog.New(handler)
}

// stores a logger in the context, used to pass request-scoped
// loggers (with trace IDs, user IDs, etc.) through the call stack.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves the logger from context.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// AddField returns a new context with an updated logger containing the new fields.
// Usage: ctx = logger.AddField(ctx, "user_id", "123")
func AddField(ctx context.Context, key string, value any) context.Context {
	l := FromContext(ctx).With(key, value)
	return WithContext(ctx, l)
}

// Call this once per request in middleware and attach to the request-scoped logger.
func RequestAttrs(method, path, requestID, ip string) []any {
	return []any{
		"request_id", requestID,
		"method", method,
		"path", path,
		"ip", ip,
	}
}

func LatencyAttr(start time.Time) slog.Attr {
	return slog.Duration("latency", time.Since(start).Round(time.Millisecond))
}
