package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
	"errors"
	"fmt"
)

type OrderRepository interface {
	GetOrderUserID(ctx context.Context, orderID string) (string, error)
	CreateOrder(ctx context.Context, userID string) (string, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	GetOrder(ctx context.Context, orderID string) (*domain.OrderInfo, error)
	GetOrdersUser(ctx context.Context, filter *domain.OrderFilter) ([]*domain.Order, error)
}

type OrderUsecase struct {
	repo OrderRepository
}

func NewOrderUsecase(repo OrderRepository) *OrderUsecase {
	return &OrderUsecase{repo: repo}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID string) (*domain.OrderInfo, error) {
	orderID, err := uc.repo.CreateOrder(ctx, userID)
	if err != nil {
		// сохраняем доменные ошибки из repository
		if errors.Is(err, domain.ErrCartEmpty) || errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		// Остальные ошибки - внутренние
		return nil, domain.ErrInternalServer
	}
	return uc.repo.GetOrder(ctx, orderID)
}

func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID, userID, status string) error {
	allowed := map[string]bool{
		"pending":    true,
		"paid":       true,
		"delivered":  true,
		"cancelled":  true,
		"on_the_way": true,
	}
	if !allowed[status] {
		return domain.ErrRequestParams
	}

	realUserID, err := uc.repo.GetOrderUserID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return err
		}
		return domain.ErrInternalServer
	}
	if realUserID != userID {
		return domain.ErrForbidden
	}

	currentOrder, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return err
		}
		return domain.ErrInternalServer
	}

	if status == "cancelled" {
		if currentOrder.Status == "pending" {
			return uc.repo.UpdateOrderStatus(ctx, orderID, "cancelled")
		}
		return fmt.Errorf("cannot cancel order in status '%s'", currentOrder.Status)
	}

	return domain.ErrForbidden
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, orderID, userID string) (*domain.OrderInfo, error) {
	realUserID, err := uc.repo.GetOrderUserID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternalServer
	}
	if realUserID != userID {
		return nil, domain.ErrForbidden
	}
	return uc.repo.GetOrder(ctx, orderID)
}

func (uc *OrderUsecase) GetOrdersUser(ctx context.Context, filter *domain.OrderFilter) ([]*domain.Order, error) {
	if filter == nil {
		return nil, domain.ErrRequestParams
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		return nil, domain.ErrRequestParams
	}
	if filter.UserID == "" {
		return nil, domain.ErrRequestParams
	}

	orders, err := uc.repo.GetOrdersUser(ctx, filter)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternalServer
	}
	return orders, nil
}

// TODO: Пофиксить валидацию UUID
// Сейчас если отправить кривой UUID типа "invalid-uuid", то будет 500 ошибка
// Надо сделать чтобы возвращалась 400 ошибка
// Просто добавить проверку uuid.Parse() в каждом хендлере перед вызовом usecase
