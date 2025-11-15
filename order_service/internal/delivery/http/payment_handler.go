package http

import (
	"apple_backend/order_service/internal/config"
	"apple_backend/order_service/internal/delivery/middlewares"
	"apple_backend/order_service/internal/delivery/transport"
	"apple_backend/order_service/internal/domain"
	"apple_backend/order_service/internal/infrastructure/yookassa"
	"apple_backend/order_service/internal/repository"
	"apple_backend/order_service/internal/usecase"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentUsecaseInterface interface {
	CreatePayment(ctx context.Context, req *domain.PaymentCreateRequest, userID string) (*domain.YookassaPaymentResponse, error)
	HandleWebhook(ctx context.Context, webhook *domain.PaymentWebhook) error
	GetPaymentByOrderID(ctx context.Context, orderID, userID string) (*domain.Payment, error)
	GetPaymentByYookassaID(ctx context.Context, yookassaID string) (*domain.Payment, error)
}

type PaymentHandler struct {
	uc             PaymentUsecaseInterface
	rs             *http_response.ResponseSender
	validator      *validator.Validate
	yookassaSecret string
}

func NewPaymentHandler(uc PaymentUsecaseInterface, secret string) *PaymentHandler {
	return &PaymentHandler{
		uc:             uc,
		rs:             http_response.NewResponseSender(logger.Global()),
		validator:      validator.New(),
		yookassaSecret: secret,
	}
}

func NewPaymentRouter(mux *http.ServeMux, db *pgxpool.Pool, config *config.Config, apiPrefix string) {
	paymentRepo := repository.NewPaymentRepoPostgres(db)
	orderRepo := repository.NewOrderRepoPostgres(db)
	yookassaClient := yookassa.NewClient(config.YookassaBaseURL, config.YookassaShopID, config.YookassaSecret)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo, orderRepo, yookassaClient)
	paymentHandler := NewPaymentHandler(paymentUC, config.YookassaSecret)

	mux.HandleFunc(apiPrefix+"payments/webhook", paymentHandler.HandleWebhook)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc(apiPrefix+"payments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			paymentHandler.CreatePayment(w, r)
		default:
			ctx := r.Context()
			log := logger.FromContext(ctx)
			log.WarnContext(ctx, "handler payments wrong method", slog.String("method", r.Method))
			paymentHandler.rs.Error(ctx, w, http.StatusMethodNotAllowed, "payments", domain.ErrHTTPMethod, nil)
		}
	})
	protectedMux.HandleFunc(apiPrefix+"payments/order/{id}", paymentHandler.GetPaymentByOrderID)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, config.JWTSecret)
	mux.Handle(apiPrefix+"payments", protectedHandler)
	mux.Handle(apiPrefix+"payments/", protectedHandler)
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler CreatePayment start")

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler CreatePayment unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "CreatePayment", domain.ErrUnauthorized, nil)
		return
	}

	var req transport.PaymentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "handler CreatePayment decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreatePayment", domain.ErrRequestParams, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		log.WarnContext(ctx, "handler CreatePayment validation failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreatePayment", domain.ErrRequestParams, err)
		return
	}

	if _, err := uuid.Parse(req.OrderID); err != nil {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreatePayment", domain.ErrRequestParams, errors.New("invalid order_id UUID"))
		return
	}

	// Проверка: платёж уже существует?
	existingPayment, err := h.uc.GetPaymentByOrderID(ctx, req.OrderID, userID)
	if err != nil {
		if !errors.Is(err, domain.ErrRowsNotFound) {
			log.ErrorContext(ctx, "handler CreatePayment get existing payment failed",
				slog.String("order_id", req.OrderID),
				slog.String("user_id", userID),
				slog.Any("err", err))
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreatePayment", domain.ErrInternalServer, err)
			return
		}
		// ErrRowsNotFound
	}

	if existingPayment != nil {
		// Разрешаем повторную оплату, если платёж НЕ успешный
		if existingPayment.Status == domain.PaymentSucceeded {
			log.WarnContext(ctx, "handler CreatePayment: payment already succeeded",
				slog.String("order_id", req.OrderID),
				slog.String("payment_id", existingPayment.ID),
				slog.String("yookassa_id", existingPayment.YookassaID))
			h.rs.Error(ctx, w, http.StatusConflict, "CreatePayment",
				errors.New("payment already succeeded"), nil)
			return
		}

		// Для canceled/pending — разрешаем создавать новый платёж
		log.InfoContext(ctx, "handler CreatePayment: existing payment is not succeeded, allowing new payment",
			slog.String("order_id", req.OrderID),
			slog.String("existing_payment_id", existingPayment.ID),
			slog.String("existing_status", string(existingPayment.Status)))
	}

	paymentCreateReq := transport.ToPaymentCreateRequest(&req)
	paymentResp, err := h.uc.CreatePayment(ctx, paymentCreateReq, userID)
	if err != nil {
		log.ErrorContext(ctx, "handler CreatePayment failed", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrRowsNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "CreatePayment", domain.ErrRowsNotFound, err)
		case errors.Is(err, domain.ErrForbidden):
			h.rs.Error(ctx, w, http.StatusForbidden, "CreatePayment", domain.ErrForbidden, err)
		case errors.Is(err, domain.ErrRequestParams):
			h.rs.Error(ctx, w, http.StatusBadRequest, "CreatePayment", domain.ErrRequestParams, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreatePayment", domain.ErrInternalServer, err)
		}
		return
	}

	payment, err := h.uc.GetPaymentByOrderID(ctx, req.OrderID, userID)
	if err != nil {
		log.ErrorContext(ctx, "handler CreatePayment get payment failed", slog.Any("err", err))
	}

	response := transport.ToPaymentResponse(payment, paymentResp)
	log.InfoContext(ctx, "handler CreatePayment success",
		slog.String("order_id", req.OrderID),
		slog.String("yookassa_id", paymentResp.ID))

	h.rs.Send(ctx, w, http.StatusCreated, response)
}

func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler HandleWebhook start")

	if r.Method != http.MethodPost {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "HandleWebhook", domain.ErrHTTPMethod, nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.rs.Error(ctx, w, http.StatusBadRequest, "HandleWebhook", domain.ErrRequestParams, err)
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-Yoo-Signature")
	if sig == "" {
		h.rs.Error(ctx, w, http.StatusUnauthorized, "HandleWebhook", errors.New("missing signature"), nil)
		return
	}

	if err := yookassa.VerifyWebhookSignature(body, sig, h.yookassaSecret); err != nil {
		log.WarnContext(ctx, "invalid webhook signature")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "HandleWebhook", errors.New("invalid signature"), nil)
		return
	}

	var webhookReq transport.PaymentWebhookRequest
	if err := json.Unmarshal(body, &webhookReq); err != nil {
		h.rs.Error(ctx, w, http.StatusBadRequest, "HandleWebhook", domain.ErrRequestParams, err)
		return
	}

	if webhookReq.Type != "notification" || webhookReq.Object.ID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "HandleWebhook", domain.ErrRequestParams, errors.New("invalid webhook"))
		return
	}

	webhook := transport.ToPaymentWebhook(&webhookReq)
	if err := h.uc.HandleWebhook(ctx, webhook); err != nil {
		log.ErrorContext(ctx, "handler HandleWebhook failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "HandleWebhook", domain.ErrInternalServer, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *PaymentHandler) GetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetPaymentByOrderID start")

	if r.Method != http.MethodGet {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetPaymentByOrderID", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetPaymentByOrderID", domain.ErrUnauthorized, nil)
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetPaymentByOrderID", domain.ErrRequestParams, errors.New("order id required"))
		return
	}

	if _, err := uuid.Parse(orderID); err != nil {
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetPaymentByOrderID", domain.ErrRequestParams, errors.New("invalid order_id UUID"))
		return
	}

	payment, err := h.uc.GetPaymentByOrderID(ctx, orderID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRowsNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "GetPaymentByOrderID", domain.ErrRowsNotFound, err)
		case errors.Is(err, domain.ErrForbidden):
			h.rs.Error(ctx, w, http.StatusForbidden, "GetPaymentByOrderID", domain.ErrForbidden, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetPaymentByOrderID", domain.ErrInternalServer, err)
		}
		return
	}

	response := transport.ToPaymentResponse(payment, nil)
	h.rs.Send(ctx, w, http.StatusOK, response)
}
