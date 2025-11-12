package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Logger = *slog.Logger

type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

var std Logger = newLogger(slog.LevelInfo)

func NewLogger(_ string, level slog.Level) Logger {
	std = newLogger(level)
	return std
}

func ContextWithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return std
	}
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(Logger); ok {
			return l
		}
	}
	return std
}

func ContextWithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type DevJSONHandler struct {
	writer *os.File
	level  slog.Level
	attrs  []slog.Attr
}

func (h *DevJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DevJSONHandler{
		writer: h.writer,
		level:  h.level,
		attrs:  append(h.attrs[:], attrs...),
	}
}

func (h *DevJSONHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *DevJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *DevJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	data := map[string]interface{}{
		"time":    r.Time.Format(time.RFC3339),
		"level":   r.Level.String(),
		"message": r.Message,
	}

	// 1. Добавляем request_id из контекста
	if rid := RequestIDFromContext(ctx); rid != "" {
		data["request_id"] = rid
	}

	// 2. Добавляем атрибуты из handler (method, url, remote_addr и т.д.)
	for _, attr := range h.attrs {
		data[attr.Key] = attr.Value.Any()
	}

	// 3. Добавляем атрибуты из записи
	r.Attrs(func(a slog.Attr) bool {
		data[a.Key] = a.Value.Any()
		return true
	})

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(h.writer, "%s\n", jsonData)
	return nil
}

func newLogger(level slog.Level) *slog.Logger {
	return slog.New(&DevJSONHandler{
		writer: os.Stdout,
		level:  level,
		attrs:  []slog.Attr{},
	})
}

func Global() Logger {
	return std
}
