package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type CartRepository interface {
	// GetCartItems метод для получения ID корзины и списка товаров
	GetCartItems(ctx context.Context, userID string) (string, []*domain.CartItem, error)

	// UpdateCartItems метод для обновления товаров в корзине
	UpdateCartItems(ctx context.Context, id string, newItems *domain.CartUpdate) error
	// DeleteCartItems метод для очистки корзины
	DeleteCartItems(ctx context.Context, id string) error
}

type CartUsecase struct {
	repo CartRepository
}

func NewCartUsecase(repo CartRepository) *CartUsecase {
	return &CartUsecase{repo: repo}
}

func (uc *CartUsecase) GetCart(ctx context.Context, id string) (*domain.Cart, error) {
	cartID, items, err := uc.repo.GetCartItems(ctx, id)
	if err != nil {
		return nil, err
	}
	cart := &domain.Cart{
		ID:    cartID,
		Items: items,
	}

	return cart, nil
}

func (uc *CartUsecase) UpdateCart(ctx context.Context, cartID string, cartUpdate *domain.CartUpdate) error {
	return uc.repo.UpdateCartItems(ctx, cartID, cartUpdate)
}

// применятся после оформления заказа
func (uc *CartUsecase) DeleteCart(ctx context.Context, cartID string) error {
	return uc.repo.DeleteCartItems(ctx, cartID)
}
