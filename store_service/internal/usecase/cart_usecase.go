package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type CartRepository interface {
	GetCartItems(ctx context.Context, userID string) ([]*domain.CartItem, error)
	UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error
	DeleteCartItems(ctx context.Context, userID string) error
}

type CartUsecase struct {
	repo CartRepository
}

func NewCartUsecase(repo CartRepository) *CartUsecase {
	return &CartUsecase{repo: repo}
}

func (uc *CartUsecase) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	items, err := uc.repo.GetCartItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domain.Cart{Items: items}, nil
}

func (uc *CartUsecase) UpdateCart(ctx context.Context, userID string, cartUpdate *domain.CartUpdate) error {
	return uc.repo.UpdateCartItems(ctx, userID, cartUpdate)
}

func (uc *CartUsecase) DeleteCart(ctx context.Context, userID string) error {
	return uc.repo.DeleteCartItems(ctx, userID)
}
