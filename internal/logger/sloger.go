package logger

import (
	"context"
	"log/slog"
	"os"
)

type SLogger struct{}

func init() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	l := slog.New(h)
	slog.SetDefault(l)
}

func NewSLogger() Logger {
	return &SLogger{}
}

func (l *SLogger) Debug(ctx context.Context, msg string, fields ...slog.Attr) {
	args := getArgs(mergeAttrs(ctx, fields))
	slog.Default().DebugContext(ctx, msg, args...)
}

func (l *SLogger) Info(ctx context.Context, msg string, fields ...slog.Attr) {
	args := getArgs(mergeAttrs(ctx, fields))
	slog.Default().InfoContext(ctx, msg, args...)
}

func (l *SLogger) Warn(ctx context.Context, msg string, fields ...slog.Attr) {
	args := getArgs(mergeAttrs(ctx, fields))
	slog.Default().WarnContext(ctx, msg, args...)
}

func (l *SLogger) Error(ctx context.Context, err error, fields ...slog.Attr) {
	args := getArgs(mergeAttrs(ctx, fields))
	slog.Default().ErrorContext(ctx, err.Error(), args...)
}
func (l *SLogger) Panic(ctx context.Context, err error, attrs ...slog.Attr) {
	l.Error(ctx, err, attrs...)
	panic(err)
}

func (l *SLogger) Fatal(ctx context.Context, err error, fields ...slog.Attr) {
	l.Error(ctx, err, fields...)
	os.Exit(1)
}

func getArgs(attrs []slog.Attr) []any {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	return args
}

func getAttrs(ctx context.Context) []slog.Attr {
	av := ctx.Value(CtxValueKey{})
	if av == nil {
		return []slog.Attr{}
	}

	res, ok := av.([]slog.Attr)
	if !ok {
		return []slog.Attr{}
	}

	return res
}

func mergeAttrs(ctx context.Context, attrs []slog.Attr) []slog.Attr {
	existing := getAttrs(ctx)
	return append(existing, attrs...)
}
