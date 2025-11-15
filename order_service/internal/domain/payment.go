package domain

import (
	"apple_backend/pkg/money"
	"time"
)

type PaymentStatus string

const (
	PaymentPending           PaymentStatus = "pending"
	PaymentWaitingForCapture PaymentStatus = "waiting_for_capture"
	PaymentSucceeded         PaymentStatus = "succeeded"
	PaymentCanceled          PaymentStatus = "canceled"
)

type Payment struct {
	ID          string                 `json:"id"`
	OrderID     string                 `json:"order_id"`
	YookassaID  string                 `json:"yookassa_id"`
	Status      PaymentStatus          `json:"status"`
	Amount      money.Money            `json:"amount"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type PaymentCreateRequest struct {
	OrderID     string      `json:"order_id"`
	Amount      money.Money `json:"amount"`
	Currency    string      `json:"currency"`
	Description string      `json:"description"`
	ReturnURL   string      `json:"return_url"`
}

type YookassaPaymentRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Capture      bool `json:"capture"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type YookassaPaymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Paid   bool   `json:"paid"`
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
	CreatedAt   string                 `json:"created_at"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Recipient   struct {
		AccountID string `json:"account_id"`
		GatewayID string `json:"gateway_id"`
	} `json:"recipient"`
	Refundable bool `json:"refundable"`
	Test       bool `json:"test"`
}

type PaymentWebhook struct {
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
