package yookassa

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/logger"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	shopID     string
	secretKey  string
	httpClient *http.Client
}

func NewClient(baseURL, shopID, secretKey string) *Client {
	return &Client{
		baseURL:   baseURL,
		shopID:    shopID,
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreatePayment(ctx context.Context, req *domain.YookassaPaymentRequest, idempotenceKey string) (*domain.YookassaPaymentResponse, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "yookassa CreatePayment start", slog.String("idempotence_key", idempotenceKey))

	url := c.baseURL + "/payments"

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.ErrorContext(ctx, "yookassa CreatePayment marshal failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.ErrorContext(ctx, "yookassa CreatePayment create request failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq, idempotenceKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.ErrorContext(ctx, "yookassa CreatePayment request failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorContext(ctx, "yookassa CreatePayment read response failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.ErrorContext(ctx, "yookassa CreatePayment API error",
			slog.Int("status", resp.StatusCode),
			slog.String("response", string(body)))
		return nil, fmt.Errorf("yookassa API error: status %d, response: %s", resp.StatusCode, string(body))
	}

	var paymentResp domain.YookassaPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		log.ErrorContext(ctx, "yookassa CreatePayment unmarshal failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.DebugContext(ctx, "yookassa CreatePayment success",
		slog.String("yookassa_id", paymentResp.ID),
		slog.String("status", paymentResp.Status))
	return &paymentResp, nil
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (*domain.YookassaPaymentResponse, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "yookassa GetPayment start", slog.String("payment_id", paymentID))

	url := c.baseURL + "/payments/" + paymentID

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.ErrorContext(ctx, "yookassa GetPayment create request failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq, uuid.NewString())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.ErrorContext(ctx, "yookassa GetPayment request failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorContext(ctx, "yookassa GetPayment read response failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.ErrorContext(ctx, "yookassa GetPayment API error",
			slog.Int("status", resp.StatusCode),
			slog.String("response", string(body)))
		return nil, fmt.Errorf("yookassa API error: status %d, response: %s", resp.StatusCode, string(body))
	}

	var paymentResp domain.YookassaPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		log.ErrorContext(ctx, "yookassa GetPayment unmarshal failed", slog.Any("err", err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.DebugContext(ctx, "yookassa GetPayment success",
		slog.String("yookassa_id", paymentResp.ID),
		slog.String("status", paymentResp.Status))
	return &paymentResp, nil
}

func (c *Client) setHeaders(req *http.Request, idempotenceKey string) {
	auth := base64.StdEncoding.EncodeToString([]byte(c.shopID + ":" + c.secretKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Idempotence-Key", idempotenceKey)
	req.Header.Set("Content-Type", "application/json")
}
