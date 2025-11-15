package usecase

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/order_service/internal/infrastructure/yookassa"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/money"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByYookassaID(ctx context.Context, yookassaID string) (*domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, yookassaID string, status domain.PaymentStatus) error
	GetByID(ctx context.Context, paymentID string) (*domain.Payment, error)
}

type PaymentUsecase struct {
	paymentRepo PaymentRepository
	orderRepo   OrderRepository
	yookassa    *yookassa.Client
}

func NewPaymentUsecase(paymentRepo PaymentRepository, orderRepo OrderRepository, yookassa *yookassa.Client) *PaymentUsecase {
	return &PaymentUsecase{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		yookassa:    yookassa,
	}
}
func (uc *PaymentUsecase) CreatePayment(ctx context.Context, req *domain.PaymentCreateRequest, userID string) (*domain.YookassaPaymentResponse, error) {
	log := logger.FromContext(ctx)

	// 1. Проверка прав
	orderUserID, err := uc.orderRepo.GetOrderUserID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if orderUserID != userID {
		return nil, domain.ErrForbidden
	}

	// 2. Проверка заказа
	order, err := uc.orderRepo.GetOrder(ctx, req.OrderID)
	if err != nil {
		return nil, domain.ErrInternalServer
	}
	if !req.Amount.Equal(money.NewMoneyFromFloat(order.Total)) {
		return nil, domain.ErrRequestParams
	}
	if order.Status != "pending" {
		return nil, fmt.Errorf("invalid order status: %s", order.Status)
	}

	// 3. Запрос к ЮKassa
	yookassaReq := &domain.YookassaPaymentRequest{
		Capture: true,
		Confirmation: struct {
			Type      string `json:"type"`
			ReturnURL string `json:"return_url"`
		}{
			Type:      "redirect",
			ReturnURL: req.ReturnURL,
		},
		Description: req.Description,
		Metadata: map[string]interface{}{
			"order_id": req.OrderID,
			"user_id":  userID,
		},
	}
	yookassaReq.Amount.Value = req.Amount.String()
	yookassaReq.Amount.Currency = req.Currency

	idempotenceKey := uuid.New().String()
	paymentResp, err := uc.yookassa.CreatePayment(ctx, yookassaReq, idempotenceKey)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	// 4. Сохраняем платёж в БД
	payment := &domain.Payment{
		ID:          uuid.New().String(),
		OrderID:     req.OrderID,
		YookassaID:  paymentResp.ID,
		Status:      domain.PaymentStatus(paymentResp.Status),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Metadata:    yookassaReq.Metadata,
	}
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		// Если БД упала — платёж в ЮKassa остался.
		// В идеале: логировать алерт + retry в фоне.
		// Для MVP — возвращаем клиенту confirmation_url (он сможет оплатить).
		log.ErrorContext(ctx, "payment saved in Yookassa but not in DB", slog.Any("err", err))
	}

	return paymentResp, nil
}

func (uc *PaymentUsecase) HandleWebhook(ctx context.Context, webhook *domain.PaymentWebhook) error {
	log := logger.FromContext(ctx)

	payment, err := uc.paymentRepo.GetByYookassaID(ctx, webhook.Object.ID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			// Платёж ещё не сохранён в БД (гонка: webhook пришёл раньше Create)
			// Логируем и выходим — вебхук пришлётся повторно через 5-10 мин.
			log.WarnContext(ctx, "payment not found in DB, webhook will retry",
				slog.String("yookassa_id", webhook.Object.ID))
			return nil
		}
		return err
	}

	newStatus := domain.PaymentStatus(webhook.Object.Status)
	if err := uc.paymentRepo.UpdateStatus(ctx, webhook.Object.ID, newStatus); err != nil {
		return err
	}

	if newStatus == domain.PaymentSucceeded {
		if err := uc.orderRepo.UpdateOrderStatus(ctx, payment.OrderID, "paid"); err != nil {
			log.ErrorContext(ctx, "failed to update order status", slog.Any("err", err))
		}
	}

	return nil
}

func (uc *PaymentUsecase) GetPaymentByOrderID(ctx context.Context, orderID, userID string) (*domain.Payment, error) {
	orderUserID, err := uc.orderRepo.GetOrderUserID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if orderUserID != userID {
		return nil, domain.ErrForbidden
	}
	return uc.paymentRepo.GetByOrderID(ctx, orderID)
}

func (uc *PaymentUsecase) GetPaymentByYookassaID(ctx context.Context, yookassaID string) (*domain.Payment, error) {
	return uc.paymentRepo.GetByYookassaID(ctx, yookassaID)
}
