package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type StoreRepository interface {
	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.Store, error)
	GetStore(ctx context.Context, id string) ([]*domain.Store, error)
	GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error)
	// CreateStore не используется на фронте
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

func (uc *StoreUsecase) GetStore(ctx context.Context, id string) (*domain.StoreAgg, error) {
	stores, err := uc.repo.GetStore(ctx, id)
	if err != nil {
		return nil, err
	}

	storeAgg := &domain.StoreAgg{
		ID:          stores[0].ID,
		Name:        stores[0].Name,
		Description: stores[0].Description,
		CityID:      stores[0].CityID,
		Address:     stores[0].Address,
		CardImg:     stores[0].CardImg,
		Rating:      stores[0].Rating,
		TagsID:      []string{stores[0].TagID},
		OpenAt:      stores[0].OpenAt,
		ClosedAt:    stores[0].ClosedAt,
	}

	if len(stores) > 1 {
		for _, store := range stores[1:] {
			storeAgg.TagsID = append(storeAgg.TagsID, store.TagID)
		}
	}

	return storeAgg, nil
}

func (uc *StoreUsecase) GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) {
	return uc.repo.GetStoreReview(ctx, id)
}

func (uc *StoreUsecase) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error) {
	if filter.Limit <= 0 {
		return nil, domain.ErrRequestParams
	}
	// допустимые поля для сортировки
	sortableFields := map[string]bool{
		"rating":    true,
		"open_at":   true,
		"closed_at": true,
	}

	if filter.Sorted != "" && !sortableFields[filter.Sorted] {
		return nil, domain.ErrRequestParams
	}

	stores, err := uc.repo.GetStores(ctx, filter)
	if err != nil {
		return nil, err
	}

	storesMap := map[string]*domain.StoreAgg{}
	for _, store := range stores {
		// если есть запись по этому товару
		if storeMap, ok := storesMap[store.ID]; ok {
			storeMap.TagsID = append(storeMap.TagsID, store.TagID)
		} else {
			storesMap[store.ID] = &domain.StoreAgg{
				ID:          store.ID,
				Name:        store.Name,
				Description: store.Description,
				CityID:      store.CityID,
				Address:     store.Address,
				CardImg:     store.CardImg,
				Rating:      store.Rating,
				TagsID:      []string{store.TagID},
				OpenAt:      store.OpenAt,
				ClosedAt:    store.ClosedAt,
			}
		}
	}

	storeList := make([]*domain.StoreAgg, 0, len(storesMap))
	for _, store := range storesMap {
		storeList = append(storeList, store)
	}

	return storeList, nil
}

func (uc *StoreUsecase) GetCities(ctx context.Context) ([]*domain.City, error) {
	return uc.repo.GetCities(ctx)
}

func (uc *StoreUsecase) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	return uc.repo.GetTags(ctx)
}
