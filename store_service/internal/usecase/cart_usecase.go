package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"

	"github.com/google/uuid"
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
	for _, item := range cartUpdate.Items {
		if _, err := uuid.Parse(item.ID); err != nil {
			return domain.ErrRequestParams // невалидный UUID
		}

		if item.Quantity < 0 {
			return domain.ErrInvalidQuantity
		}
	}

	err := uc.repo.UpdateCartItems(ctx, userID, cartUpdate)
	if err != nil {
		return err
	}
	return nil
}

func (uc *CartUsecase) DeleteCart(ctx context.Context, userID string) error {
	return uc.repo.DeleteCartItems(ctx, userID)
}

// TODO: Пофиксить валидацию UUID
// Сейчас если отправить кривой UUID типа "invalid-uuid", то будет 500 ошибка
// Надо сделать чтобы возвращалась 400 ошибка
// Просто добавить проверку uuid.Parse() в каждом хендлере перед вызовом usecase
