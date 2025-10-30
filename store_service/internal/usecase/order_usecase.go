package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type OrderRepository interface {
	// CreateOrder создает новый заказ пользователя
	CreateOrder(ctx context.Context, userID string) (string, error)
	// UpdateOrderStatus обновляет статус заказа
	UpdateOrderStatus(ctx context.Context, id, status string) error
	// GetOrder получить информацию о заказе по ID
	GetOrder(ctx context.Context, id string) (*domain.OrderInfo, error)
	// GetOrdersUser получить все заказы пользователя
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

	orderInfo, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return orderInfo, nil
}

func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, id, status string) error {
	statuses := map[string]bool{
		"pending":    true,
		"paid":       true,
		"delivered":  true,
		"cancelled":  true,
		"on the way": true,
	}
	if !statuses[status] {
		return domain.ErrRequestParams
	}

	return uc.repo.UpdateOrderStatus(ctx, id, status)
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, id string) (*domain.OrderInfo, error) {
	order, err := uc.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUsecase) GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	orders, err := uc.repo.GetOrdersUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
