package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/transport"
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/repository"
	"apple_backend/store_service/internal/usecase"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/goccy/go-json"

	"github.com/google/uuid"
)

type StoreUsecaseInterface interface {
	GetCities(ctx context.Context) ([]*domain.City, error)

	GetStore(ctx context.Context, id string) (*domain.StoreAgg, error)
	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error)
	CreateStore(ctx context.Context, name, description, cityID, address, cardImg, openAt, closedAt string, rating float64) error

	GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) // TODO Вынести

	GetTags(ctx context.Context) ([]*domain.StoreTag, error)
	GetCategories(ctx context.Context) ([]*domain.Category, error)

	SearchStoresWithItems(ctx context.Context, filter *domain.StoreSearchFilter) ([]*domain.StoreWithItems, error)
}

type StoreHandler struct {
	uc StoreUsecaseInterface
	rs *http_response.ResponseSender
}

func NewStoreHandler(uc StoreUsecaseInterface) *StoreHandler {
	return &StoreHandler{
		uc: uc,
		rs: http_response.NewResponseSender(logger.Global()),
	}
}

func NewStoreRouter(
	mux *http.ServeMux,
	db repository.PgxIface,
	embeddingClient interface {
		GetEmbedding(ctx context.Context, text string) ([]float32, error)
		GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error)
	},
	apiPrefix string,
) {
	storeRepo := repository.NewStoreRepoPostgres(db)
	storeUC := usecase.NewStoreUsecase(storeRepo, embeddingClient)
	storeHandler := NewStoreHandler(storeUC)

	mux.HandleFunc(apiPrefix+"stores/search/items", storeHandler.SearchStoresWithItems)
	mux.HandleFunc(apiPrefix+"stores/cities", storeHandler.GetCities)
	mux.HandleFunc(apiPrefix+"stores/tags", storeHandler.GetTags)
	mux.HandleFunc(apiPrefix+"stores/categories", storeHandler.GetCategories)
	mux.HandleFunc(apiPrefix+"stores/{id}/reviews", storeHandler.GetStoreReview) // TODO: Вынести
	mux.HandleFunc(apiPrefix+"stores/{id}", storeHandler.GetStore)
	mux.HandleFunc(apiPrefix+"stores", storeHandler.GetStores)
}

func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler CreateStore start")

	if r.Method != http.MethodPost {
		log.WarnContext(ctx, "handler CreateStore wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "CreateStore", domain.ErrHTTPMethod, nil)
		return
	}

	req := &domain.Store{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		log.ErrorContext(ctx, "handler CreateStore decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateStore", domain.ErrRequestParams, err)
		return
	}

	err := h.uc.CreateStore(ctx, req.Name, req.Description, req.CityID, req.Address, req.CardImg, req.OpenAt,
		req.ClosedAt, req.Rating)
	if err != nil {
		log.ErrorContext(ctx, "handler CreateStore usecase failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrStoreExist) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "CreateStore", domain.ErrStoreExist, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "CreateStore", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler CreateStore success")
	w.WriteHeader(http.StatusCreated)
}

func (h *StoreHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetStore start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetStore wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetStore", domain.ErrHTTPMethod, nil)
		return
	}

	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		log.WarnContext(ctx, "handler GetStore invalid id", slog.String("id", id))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetStore", domain.ErrRequestParams, nil)
		return
	}

	store, err := h.uc.GetStore(ctx, id)
	if err != nil {
		log.ErrorContext(ctx, "handler GetStore usecase failed", slog.Any("err", err), slog.String("id", id))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetStore", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetStore", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetStore success", slog.String("id", id))
	responseStore := transport.ToStoreResponse(store)
	h.rs.Send(ctx, w, http.StatusOK, responseStore)
}

func (h *StoreHandler) GetStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetStores start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetStores wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetStores", domain.ErrHTTPMethod, nil)
		return
	}

	q := r.URL.Query()
	limitStr := q.Get("limit")
	if limitStr == "" {
		log.WarnContext(ctx, "handler GetStores missing limit")
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetStores", domain.ErrRequestParams, errors.New("limit required"))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		log.WarnContext(ctx, "handler GetStores invalid limit", slog.String("limit", limitStr))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetStores", domain.ErrRequestParams, errors.New("invalid limit"))
		return
	}

	tagIDs := q["tag_id"]
	if tagIDs == nil {
		tagIDs = []string{}
	}

	categoryIDs := q["category_id"]
	if categoryIDs == nil {
		categoryIDs = []string{}
	}

	filter := &domain.StoreFilter{
		Limit:       limit,
		LastID:      q.Get("last_id"),
		TagIDs:      tagIDs,
		CategoryIDs: categoryIDs,
		CityID:      q.Get("city_id"),
		Search:      q.Get("search"),
		Sorted:      q.Get("sorted"),
		Desc:        q.Has("desc") && q.Get("desc") == "true",
	}

	stores, err := h.uc.GetStores(ctx, filter)
	if err != nil {
		log.ErrorContext(ctx, "handler GetStores usecase failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrRequestParams) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "GetStores", domain.ErrRequestParams, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetStores", domain.ErrInternalServer, err)
		return
	}

	for _, s := range stores {
		if s.CardImg != "" {
			s.CardImg = "/images/stores/" + s.CardImg
		}
	}

	log.InfoContext(ctx, "handler GetStores success", slog.Int("count", len(stores)))
	responseStores := transport.ToStoreResponses(stores)
	h.rs.Send(ctx, w, http.StatusOK, responseStores)
}

func (h *StoreHandler) SearchStoresWithItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler SearchStoresWithItems start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler SearchStoresWithItems wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "SearchStoresWithItems", domain.ErrHTTPMethod, nil)
		return
	}

	q := r.URL.Query()
	limitStr := q.Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		log.WarnContext(ctx, "handler SearchStoresWithItems invalid limit", slog.String("limit", limitStr))
		h.rs.Error(ctx, w, http.StatusBadRequest, "SearchStoresWithItems", domain.ErrRequestParams, errors.New("invalid limit"))
		return
	}

	minPriceStr := q.Get("min_price")
	minPrice := 0.0
	if minPriceStr != "" {
		if p, err := strconv.ParseFloat(minPriceStr, 64); err == nil && p >= 0 {
			minPrice = p
		}
	}

	maxPriceStr := q.Get("max_price")
	maxPrice := 999999.0
	if maxPriceStr != "" {
		if p, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && p > 0 {
			maxPrice = p
		}
	}
	tagIDs := q["tag_id"]
	if tagIDs == nil {
		tagIDs = []string{}
	}

	categoryIDs := q["category_id"]
	if categoryIDs == nil {
		categoryIDs = []string{}
	}

	itemTypes := q["item_type"]
	if itemTypes == nil {
		itemTypes = []string{}
	}

	filter := &domain.StoreSearchFilter{
		Search:      q.Get("search"),
		TagIDs:      tagIDs,
		CategoryIDs: categoryIDs,
		CityID:      q.Get("city_id"),
		ItemTypes:   itemTypes,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		Limit:       limit,
		LastID:      q.Get("last_id"),
	}

	stores, err := h.uc.SearchStoresWithItems(ctx, filter)
	if err != nil {
		log.ErrorContext(ctx, "handler SearchStoresWithItems usecase failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrRequestParams) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "SearchStoresWithItems", domain.ErrRequestParams, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "SearchStoresWithItems", domain.ErrInternalServer, err)
		return
	}

	for _, s := range stores {
		if s.CardImg != "" {
			s.CardImg = "/images/stores/" + s.CardImg
		}
		for _, item := range s.Items {
			if item.CardImg != "" {
				item.CardImg = "/images/items/" + item.CardImg
			}
		}
	}

	log.InfoContext(ctx, "handler SearchStoresWithItems success", slog.Int("count", len(stores)))
	responseStores := transport.ToStoreWithItemsResponses(stores)
	h.rs.Send(ctx, w, http.StatusOK, responseStores)
}

func (h *StoreHandler) GetStoreReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetStoreReview start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetStoreReview wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetStoreReview", domain.ErrHTTPMethod, nil)
		return
	}

	storeID := r.PathValue("id")
	if _, err := uuid.Parse(storeID); err != nil {
		log.WarnContext(ctx, "handler GetStoreReview invalid id", slog.String("id", storeID))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetStoreReview", domain.ErrRequestParams, nil)
		return
	}

	reviews, err := h.uc.GetStoreReview(ctx, storeID)
	if err != nil {
		log.ErrorContext(ctx, "handler GetStoreReview usecase failed", slog.Any("err", err), slog.String("store_id", storeID))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetStoreReview", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetStoreReview", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetStoreReview success",
		slog.String("store_id", storeID),
		slog.Int("reviews_count", len(reviews)))
	responseReview := transport.ToStoreReviews(reviews)
	h.rs.Send(ctx, w, http.StatusOK, responseReview)
}

func (h *StoreHandler) GetCities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetCities start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetCities wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetCities", domain.ErrHTTPMethod, nil)
		return
	}

	cities, err := h.uc.GetCities(ctx)
	if err != nil {
		log.ErrorContext(ctx, "handler GetCities usecase failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetCities", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetCities", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetCities success", slog.Int("count", len(cities)))
	responseCities := transport.ToCityResponses(cities)
	h.rs.Send(ctx, w, http.StatusOK, responseCities)
}

func (h *StoreHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetTags start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetTags wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetTags", domain.ErrHTTPMethod, nil)
		return
	}

	tags, err := h.uc.GetTags(ctx)
	if err != nil {
		log.ErrorContext(ctx, "handler GetTags usecase failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetTags", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetTags", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetTags success", slog.Int("count", len(tags)))
	responseTags := transport.ToTagResponses(tags)
	h.rs.Send(ctx, w, http.StatusOK, responseTags)
}

func (h *StoreHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetCategories start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetCategories wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetCategories", domain.ErrHTTPMethod, nil)
		return
	}

	categories, err := h.uc.GetCategories(ctx)
	if err != nil {
		log.ErrorContext(ctx, "handler GetCategories usecase failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetCategories", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetCategories", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetCategories success", slog.Int("count", len(categories)))
	responseCategories := transport.ToCategoryResponses(categories)
	h.rs.Send(ctx, w, http.StatusOK, responseCategories)
}
