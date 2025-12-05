package usecase

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"log/slog"
)

type StoreRepository interface {
	GetCities(ctx context.Context) ([]*domain.City, error)

	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error)
	GetStore(ctx context.Context, id string) (*domain.StoreAgg, error)
	CreateStore(ctx context.Context, store *domain.Store) error

	GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) // TODO Вынести

	GetTags(ctx context.Context) ([]*domain.StoreTag, error)
	GetCategories(ctx context.Context) ([]*domain.Category, error)

	SearchStoresWithItems(ctx context.Context, filter *domain.StoreSearchFilter) ([]*domain.StoreWithItems, error)
	SearchStoresHybrid(ctx context.Context, filter *domain.StoreSearchFilter, embedding []float32) ([]*domain.StoreWithItems, error)
}

type StoreUsecase struct {
	repo            StoreRepository
	embeddingClient domain.EmbeddingClient
}

func NewStoreUsecase(repo StoreRepository, embeddingClient domain.EmbeddingClient) *StoreUsecase {
	return &StoreUsecase{
		repo:            repo,
		embeddingClient: embeddingClient,
	}
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

func (uc *StoreUsecase) SearchStoresWithItems(ctx context.Context, filter *domain.StoreSearchFilter) ([]*domain.StoreWithItems, error) {
	log := logger.FromContext(ctx)

	if filter.Limit <= 0 {
		return nil, domain.ErrRequestParams
	}
	if filter.MaxPrice < filter.MinPrice {
		return nil, domain.ErrRequestParams
	}

	if filter.Search == "" {
		return uc.repo.SearchStoresWithItems(ctx, filter)
	}

	// Гибридный поиск через embedding-сервис
	embedding, err := uc.embeddingClient.GetEmbedding(ctx, filter.Search)
	if err != nil {
		log.WarnContext(ctx, "embedding service unavailable, fallback to traditional search",
			slog.Any("err", err),
			slog.String("search", filter.Search),
		)
		// Fallback на обычный поиск
		return uc.repo.SearchStoresWithItems(ctx, filter)
	}

	log.DebugContext(ctx, "using hybrid search",
		slog.String("search", filter.Search),
		slog.Int("embedding_dim", len(embedding)),
	)

	return uc.repo.SearchStoresHybrid(ctx, filter, embedding)
}

func (uc *StoreUsecase) GetCities(ctx context.Context) ([]*domain.City, error) {
	return uc.repo.GetCities(ctx)
}

func (uc *StoreUsecase) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	return uc.repo.GetTags(ctx)
}

func (uc *StoreUsecase) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	return uc.repo.GetCategories(ctx)
}
