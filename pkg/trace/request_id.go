package trace

import "context"

type ctxKey string

const RequestIDKey ctxKey = "request_id"

// GetRequestID извлечение Request ID из контекста
func GetRequestID(ctx context.Context) string {
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
