package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

// RequestIDFromCtx извлечение Request ID из контекста
func RequestIDFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// SetRequestID добавляет Request ID в контекст, если нету
func SetRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, reqID)
}

type Logger struct {
	logger *slog.Logger
}

func NewLogger(logPath string, level slog.Level) *Logger {
	dir, _ := filepath.Split(logPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		slog.Error("Не удалось создать директорию для %s: %s", logPath, err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Не удалось создать лог файл %s: %s", logPath, err)
	}

	handler := slog.NewTextHandler(
		file,
		&slog.HandlerOptions{Level: level},
	)

	return &Logger{
		logger: slog.New(handler),
	}
}

func NewNilLogger() *Logger {
	handler := slog.NewTextHandler(
		io.Discard,
		&slog.HandlerOptions{Level: slog.LevelError},
	)

	return &Logger{
		logger: slog.New(handler),
	}
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, meta map[string]interface{}) {
	attrs := []slog.Attr{}
	if ctx != nil {
		if reqID := RequestIDFromCtx(ctx); reqID != "" {
			attrs = append(attrs, slog.String("request_id", reqID))
		}
	}
	for k, v := range meta {
		attrs = append(attrs, slog.Any(k, v))
	}
	l.logger.LogAttrs(ctx, level, msg, attrs...)
}

func (l *Logger) Error(ctx context.Context, message string, meta map[string]interface{}) {
	l.log(ctx, slog.LevelError, message, meta)
}

func (l *Logger) Warn(ctx context.Context, message string, meta map[string]interface{}) {
	l.log(ctx, slog.LevelWarn, message, meta)
}

func (l *Logger) Info(ctx context.Context, message string, meta map[string]interface{}) {
	l.log(ctx, slog.LevelInfo, message, meta)
}

func (l *Logger) Debug(ctx context.Context, message string, meta map[string]interface{}) {
	l.log(ctx, slog.LevelDebug, message, meta)
}
