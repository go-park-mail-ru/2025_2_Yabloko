package usecase

import (
	"apple_backend/order_service/internal/domain"
	"context"
	"errors"
	"time"
)

type PromoRepository interface {
	GetActiveForUserByCode(ctx context.Context, userID, code string, now time.Time) (*domain.Promocode, error)
	MarkUsed(ctx context.Context, userID, promoID string) error
}

type PromoUsecase struct {
	repo PromoRepository
}

func NewPromoUsecase(repo PromoRepository) *PromoUsecase {
	return &PromoUsecase{repo: repo}
}

func (uc *PromoUsecase) CheckPromo(ctx context.Context, userID, code string) (*domain.PromoCheckResult, error) {
	now := time.Now()
	p, err := uc.repo.GetActiveForUserByCode(ctx, userID, code, now)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return nil, domain.ErrRowsNotFound
		}
		return nil, domain.ErrInternalServer
	}

	return &domain.PromoCheckResult{
		RelativeDiscount: p.RelativeDiscount,
		AbsoluteDiscount: p.AbsoluteDiscount,
	}, nil
}
