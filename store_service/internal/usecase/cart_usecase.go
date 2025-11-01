package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type CartRepository interface {
	// GetCartItems метод для получения ID корзины и списка товаров
	GetCartItems(ctx context.Context, userID string) ([]*domain.CartItem, error)

	// UpdateCartItems метод для обновления товаров в корзине
	UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error
	// DeleteCartItems метод для очистки корзины
	DeleteCartItems(ctx context.Context, userID string) error
}

type CartUsecase struct {
	repo CartRepository
}

func NewCartUsecase(repo CartRepository) *CartUsecase {
	return &CartUsecase{repo: repo}
}

func (uc *CartUsecase) GetCart(ctx context.Context, id string) (*domain.Cart, error) {
	items, err := uc.repo.GetCartItems(ctx, id)
	if err != nil {
		return nil, err
	}
	cart := &domain.Cart{
		Items: items,
	}

	return cart, nil
}

func (uc *CartUsecase) UpdateCart(ctx context.Context, userID string, cartUpdate *domain.CartUpdate) error {
	return uc.repo.UpdateCartItems(ctx, userID, cartUpdate)
}

// применятся после оформления заказа (если будем распиливать)
func (uc *CartUsecase) DeleteCart(ctx context.Context, userID string) error {
	return uc.repo.DeleteCartItems(ctx, userID)
}
