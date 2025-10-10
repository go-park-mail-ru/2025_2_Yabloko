package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type StoreRepository interface {
	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.Store, error)
	GetStore(ctx context.Context, id string) (*domain.Store, error)
	// CreateStore не используется на фронте
	CreateStore(ctx context.Context, store *domain.Store) error
}

type StoreUsecase struct {
	repo StoreRepository
}

func NewStoreUsecase(repo StoreRepository) *StoreUsecase {
	return &StoreUsecase{repo: repo}
}

func (uc *StoreUsecase) CreateStore(ctx context.Context,
	name, description, cityID, address, cardImg, openAt, closedAt string,
	rating float64) error {
	store := &domain.Store{
		Name:        name,
		Description: description,
		CityID:      cityID,
		Address:     address,
		CardImg:     cardImg,
		OpenAt:      openAt,
		ClosedAt:    closedAt,
		Rating:      rating,
	}

	err := uc.repo.CreateStore(ctx, store)
	if err != nil {
		return err
	}

	return nil
}

func (uc *StoreUsecase) GetStore(ctx context.Context, id string) (*domain.Store, error) {
	return uc.repo.GetStore(ctx, id)
}

func (uc *StoreUsecase) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.Store, error) {
	return uc.repo.GetStores(ctx, filter)
}
