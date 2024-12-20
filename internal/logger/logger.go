package logger

import (
	"context"
	"log/slog"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...slog.Attr)
	Info(ctx context.Context, msg string, fields ...slog.Attr)
	Warn(ctx context.Context, msg string, fields ...slog.Attr)
	Error(ctx context.Context, err error, fields ...slog.Attr)
	Panic(ctx context.Context, err error, attrs ...slog.Attr)
	Fatal(ctx context.Context, err error, fields ...slog.Attr)
}
type CtxValueKey struct{}

var NewLogger = func() Logger {
	return NewSLogger()
}
