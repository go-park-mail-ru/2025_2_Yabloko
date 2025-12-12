package http

import (
	"apple_backend/order_service/internal/delivery/transport"
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/middlewares"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

const apiV0Prefix = "/api/v0/"

type PromoUsecaseInterface interface {
	CheckPromo(ctx context.Context, userID, code string) (*domain.PromoCheckResult, error)
}

type PromoHandler struct {
	uc        PromoUsecaseInterface
	rs        *http_response.ResponseSender
	validator *validator.Validate
}

func NewPromoHandler(uc PromoUsecaseInterface) *PromoHandler {
	return &PromoHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(logger.Global()),
		validator: validator.New(),
	}
}

func NewPromoRouter(mux *http.ServeMux, handler *PromoHandler, apiPrefix string) {
	mux.HandleFunc(apiV0Prefix+"promo/check", handler.Check)
}

func (h *PromoHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler Promo.Check start")

	if r.Method != http.MethodPost {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "PromoCheck", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler Promo.Check unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "PromoCheck", domain.ErrUnauthorized, nil)
		return
	}

	var req transport.PromoCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "handler Promo.Check decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "PromoCheck", domain.ErrRequestParams, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		log.WarnContext(ctx, "handler Promo.Check validation failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "PromoCheck", domain.ErrRequestParams, err)
		return
	}

	res, err := h.uc.CheckPromo(ctx, userID, req.Promo)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "PromoCheck", domain.ErrRequestParams, errors.New("invalid or used promo"))
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "PromoCheck", domain.ErrInternalServer, err)
		return
	}

	resp := transport.PromoCheckResponse{
		RelativeDiscount: res.RelativeDiscount,
	}
	h.rs.Send(ctx, w, http.StatusOK, resp)
}
