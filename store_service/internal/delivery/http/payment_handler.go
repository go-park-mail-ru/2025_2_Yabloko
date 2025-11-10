package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/payment"
	"net/http"

	"github.com/google/uuid"
)

type PaymentHandler struct {
	rs *http_response.ResponseSender
}

func NewPaymentHandler(log logger.Logger) *PaymentHandler {
	return &PaymentHandler{
		rs: http_response.NewResponseSender(log),
	}
}

func (h *PaymentHandler) FakePayment(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	returnURL := r.URL.Query().Get("return_url")
	price := r.URL.Query().Get("price") // опциональный параметр

	if orderID == "" || returnURL == "" {
		http.Error(w, "Missing order_id or return_url", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(orderID); err != nil {
		http.Error(w, "Invalid order_id", http.StatusBadRequest)
		return
	}

	data := &payment.PaymentPageData{
		OrderID:      orderID,
		OrderIDShort: orderID[:8],
		ReturnURL:    returnURL,
		Price:        price, // будет использована мок цена если пустая
	}

	if err := payment.RenderPaymentPage(w, data); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
