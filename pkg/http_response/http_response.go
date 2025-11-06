package http_response

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/trace"
	"context"
	"encoding/json"
	"net/http"
)

type ErrResponse struct {
	Err string `json:"error"`
}

type ResponseSender struct {
	log logger.Logger
}

func NewResponseSender(log logger.Logger) *ResponseSender {
	return &ResponseSender{log: log}
}

// универсальная отправка http ответа
func (rs *ResponseSender) Send(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	requestID := trace.GetRequestID(ctx)

	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}

	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// универсальная отправка http ошибки
func (rs *ResponseSender) Error(ctx context.Context, w http.ResponseWriter, statusCode int,
	errMessage string, userErr error, internalErr error) {
	if internalErr != nil {
		rs.log.Error(errMessage, map[string]interface{}{"userErr": userErr, "internalErr": internalErr})
	} else {
		rs.log.Warn(errMessage, map[string]interface{}{"userErr": userErr})
	}

	resp := ErrResponse{
		Err: userErr.Error(),
	}

	rs.Send(ctx, w, statusCode, resp)
}
