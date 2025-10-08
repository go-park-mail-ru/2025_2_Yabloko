package http_response

import (
	"apple_backend/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
)

type ErrResponse struct {
	Err       string `json:"error"`
	RequestID string `json:"request_id"`
}

type ResponseSender struct {
	log *logger.Logger
}

func NewResponseSender(log *logger.Logger) *ResponseSender {
	return &ResponseSender{log: log}
}

func (rs *ResponseSender) Send(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	requestID := logger.GetRequestID(ctx)

	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}

	w.WriteHeader(statusCode)

	if data != nil {
		switch v := data.(type) {
		case map[string]interface{}:
			v["request_id"] = requestID
		case map[string]string:
			v["request_id"] = requestID
		}

		json.NewEncoder(w).Encode(data)
	}
}

func (rs *ResponseSender) Error(ctx context.Context, w http.ResponseWriter, statusCode int,
	errMessage string, userErr error, internalErr error) {
	requestID := logger.GetRequestID(ctx)

	if internalErr != nil {
		rs.log.Error(ctx, errMessage, map[string]interface{}{"userErr": userErr, "internalErr": internalErr})
	} else {
		rs.log.Warn(ctx, errMessage, map[string]interface{}{"userErr": userErr})
	}

	resp := ErrResponse{
		Err:       userErr.Error(),
		RequestID: requestID,
	}

	rs.Send(ctx, w, statusCode, resp)
}
