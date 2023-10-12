package cslog

import (
	"context"

	"log/slog"
)

var slogkey = struct{}{}

// Ctx returns the logger associated with the context or the default Logger if absent
func Ctx(ctx context.Context) *slog.Logger {
	l := ctx.Value(slogkey)
	if l == nil {
		return slog.Default()
	}
	return l.(*slog.Logger)
}

// WithLogger returns a new context with the given logger attached
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, slogkey, l)
}
