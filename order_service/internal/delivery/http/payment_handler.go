package http

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/payment"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type PaymentHandler struct {
	rs *http_response.ResponseSender
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		rs: http_response.NewResponseSender(logger.Global()),
	}
}

func (h *PaymentHandler) FakePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler FakePayment start")

	orderID := r.URL.Query().Get("order_id")
	returnURL := r.URL.Query().Get("return_url")
	price := r.URL.Query().Get("price") // опциональный параметр

	log.DebugContext(ctx, "handler FakePayment parameters",
		slog.String("order_id", orderID),
		slog.String("return_url", returnURL),
		slog.String("price", price))

	if orderID == "" || returnURL == "" {
		log.WarnContext(ctx, "handler FakePayment missing required parameters",
			slog.String("order_id", orderID),
			slog.String("return_url", returnURL))
		h.rs.Error(ctx, w, http.StatusBadRequest, "FakePayment", domain.ErrRequestParams, nil)
		return
	}

	if _, err := uuid.Parse(orderID); err != nil {
		log.WarnContext(ctx, "handler FakePayment invalid order_id",
			slog.String("order_id", orderID),
			slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "FakePayment", domain.ErrRequestParams, nil)
		return
	}

	data := &payment.PaymentPageData{
		OrderID:      orderID,
		OrderIDShort: orderID[:8],
		ReturnURL:    returnURL,
		Price:        price, // мок-сервис использует значение или дефолт
	}

	log.DebugContext(ctx, "handler FakePayment rendering payment page",
		slog.String("order_id_short", data.OrderIDShort),
		slog.String("price", data.Price))

	if err := payment.RenderPaymentPage(w, data); err != nil {
		log.ErrorContext(ctx, "handler FakePayment render failed",
			slog.Any("err", err),
			slog.String("order_id", orderID))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "FakePayment", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler FakePayment success",
		slog.String("order_id", orderID),
		slog.String("return_url", returnURL))
}
