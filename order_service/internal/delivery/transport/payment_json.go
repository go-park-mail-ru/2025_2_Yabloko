package transport

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/money"
	"time"
)

type PaymentCreateRequest struct {
	OrderID     string `json:"order_id" validate:"required,uuid4"`
	Amount      string `json:"amount" validate:"required"`
	Currency    string `json:"currency" validate:"required,oneof=RUB USD EUR"`
	Description string `json:"description" validate:"max=128"`
	ReturnURL   string `json:"return_url" validate:"required,url"`
}

type PaymentResponse struct {
	ID           string        `json:"id"`
	OrderID      string        `json:"order_id"`
	YookassaID   string        `json:"yookassa_id"`
	Status       string        `json:"status"`
	Amount       string        `json:"amount"`
	Currency     string        `json:"currency"`
	Description  string        `json:"description"`
	Confirmation *Confirmation `json:"confirmation,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Confirmation struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

type PaymentWebhookRequest struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Paid   bool   `json:"paid"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		CreatedAt   string                 `json:"created_at"`
		Description string                 `json:"description"`
		Metadata    map[string]interface{} `json:"metadata"`
		Recipient   struct {
			AccountID string `json:"account_id"`
			GatewayID string `json:"gateway_id"`
		} `json:"recipient"`
		Refundable bool `json:"refundable"`
		Test       bool `json:"test"`
	} `json:"object"`
}

func ToPaymentResponse(payment *domain.Payment, confirmation *domain.YookassaPaymentResponse) *PaymentResponse {
	resp := &PaymentResponse{
		ID:          payment.ID,
		OrderID:     payment.OrderID,
		YookassaID:  payment.YookassaID,
		Status:      string(payment.Status),
		Amount:      payment.Amount.String(),
		Currency:    payment.Currency,
		Description: payment.Description,
		CreatedAt:   payment.CreatedAt,
		UpdatedAt:   payment.UpdatedAt,
	}
	if confirmation != nil && confirmation.Confirmation.ConfirmationURL != "" {
		resp.Confirmation = &Confirmation{
			Type:            confirmation.Confirmation.Type,
			ConfirmationURL: confirmation.Confirmation.ConfirmationURL,
		}
	}
	return resp
}

func ToPaymentCreateRequest(req *PaymentCreateRequest) *domain.PaymentCreateRequest {
	amount, _ := money.NewMoneyFromString(req.Amount)
	return &domain.PaymentCreateRequest{
		OrderID:     req.OrderID,
		Amount:      amount,
		Currency:    req.Currency,
		Description: req.Description,
		ReturnURL:   req.ReturnURL,
	}
}

func ToPaymentWebhook(req *PaymentWebhookRequest) *domain.PaymentWebhook {
	webhook := &domain.PaymentWebhook{
		Type:  req.Type,
		Event: req.Event,
	}
	webhook.Object.ID = req.Object.ID
	webhook.Object.Status = req.Object.Status
	webhook.Object.Paid = req.Object.Paid
	webhook.Object.Amount.Value = req.Object.Amount.Value
	webhook.Object.Amount.Currency = req.Object.Amount.Currency
	webhook.Object.CreatedAt = req.Object.CreatedAt
	webhook.Object.Description = req.Object.Description
	webhook.Object.Metadata = req.Object.Metadata
	webhook.Object.Recipient.AccountID = req.Object.Recipient.AccountID
	webhook.Object.Recipient.GatewayID = req.Object.Recipient.GatewayID
	webhook.Object.Refundable = req.Object.Refundable
	webhook.Object.Test = req.Object.Test
	return webhook
}
