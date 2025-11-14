package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/payment"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type FakePaymentHandler struct {
	rs *http_response.ResponseSender
}

func NewFakePaymentHandler() *FakePaymentHandler {
	return &FakePaymentHandler{
		rs: http_response.NewResponseSender(logger.Global()),
	}
}

func (h *FakePaymentHandler) FakePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler FakePayment start")

	orderID := r.URL.Query().Get("order_id")
	returnURL := r.URL.Query().Get("return_url")

	if orderID == "" || returnURL == "" {
		h.badRequest(ctx, w, "missing order_id or return_url")
		return
	}

	if _, err := uuid.Parse(orderID); err != nil {
		h.badRequest(ctx, w, "invalid order_id UUID")
		return
	}

	data := &payment.PaymentPageData{
		OrderID:      orderID,
		OrderIDShort: orderID[:8],
		ReturnURL:    returnURL,
	}

	if err := payment.RenderPaymentPage(w, data); err != nil {
		log.ErrorContext(ctx, "FakePayment render failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "FakePayment", errors.New("render failed"), nil)
		return
	}

	log.InfoContext(ctx, "FakePayment success", slog.String("order_id", orderID))
}

func (h *FakePaymentHandler) badRequest(ctx context.Context, w http.ResponseWriter, msg string) {
	logger.FromContext(ctx).WarnContext(ctx, "FakePayment bad request", slog.String("msg", msg))
	h.rs.Error(ctx, w, http.StatusBadRequest, "FakePayment", errors.New(msg), nil)
}
