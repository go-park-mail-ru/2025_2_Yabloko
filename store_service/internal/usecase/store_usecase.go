package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type StoreRepository interface {
	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error)
	GetStore(ctx context.Context, id string) (*domain.StoreAgg, error)
	GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error)
	CreateStore(ctx context.Context, store *domain.Store) error
	GetCities(ctx context.Context) ([]*domain.City, error)
	GetTags(ctx context.Context) ([]*domain.StoreTag, error)
}

type StoreUsecase struct {
	repo StoreRepository
}

func NewStoreUsecase(repo StoreRepository) *StoreUsecase {
	return &StoreUsecase{repo: repo}
}

func (uc *StoreUsecase) CreateStore(ctx context.Context,
	name, description, cityID, address, cardImg, openAt, closedAt string, rating float64) error {
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
	return uc.repo.CreateStore(ctx, store)
}

func (uc *StoreUsecase) GetStore(ctx context.Context, id string) (*domain.StoreAgg, error) {
	return uc.repo.GetStore(ctx, id)
}

func (uc *StoreUsecase) GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) {
	return uc.repo.GetStoreReview(ctx, id)
}

func (uc *StoreUsecase) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error) {
	if filter.Limit <= 0 {
		return nil, domain.ErrRequestParams
	}
	sortable := map[string]bool{"rating": true, "open_at": true, "closed_at": true}
	if filter.Sorted != "" && !sortable[filter.Sorted] {
		return nil, domain.ErrRequestParams
	}

	stores, err := uc.repo.GetStores(ctx, filter)
	if err != nil {
		return nil, err
	}

	return stores, nil
}

func (uc *StoreUsecase) GetCities(ctx context.Context) ([]*domain.City, error) {
	return uc.repo.GetCities(ctx)
}

func (uc *StoreUsecase) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	return uc.repo.GetTags(ctx)
}
