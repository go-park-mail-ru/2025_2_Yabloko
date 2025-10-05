package usecase

import (
	"apple_backend/store_service/internal/domain"
)

type StoreUsecase struct {
	repo domain.StoreRepository
}

func NewStoreUsecase(repo domain.StoreRepository) *StoreUsecase {
	return &StoreUsecase{repo: repo}
}

func (uc *StoreUsecase) CreateStore(name, description, cityID, address, cardImg, openAt, closedAt string,
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

	if err := uc.repo.CreateStore(store); err != nil {
		return err
	}

	return nil
}

func (uc *StoreUsecase) GetStore(id string) (*domain.Store, error) {
	return uc.repo.GetStore(id)
}

func (uc *StoreUsecase) GetStores(filter *domain.StoreFilter) ([]*domain.Store, error) {
	return uc.repo.GetStores(filter)
}
