package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
	"fmt"
)

type OrderRepository interface {
	GetOrderUserID(ctx context.Context, orderID string) (string, error)
	CreateOrder(ctx context.Context, userID string) (string, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	GetOrder(ctx context.Context, orderID string) (*domain.OrderInfo, error)
	GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error)
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
		return nil, err
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
		return err
	}
	if realUserID != userID {
		return domain.ErrForbidden
	}
	currentOrder, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
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
		return nil, err
	}
	if realUserID != userID {
		return nil, domain.ErrForbidden
	}
	return uc.repo.GetOrder(ctx, orderID)
}

func (uc *OrderUsecase) GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	return uc.repo.GetOrdersUser(ctx, userID)
}
