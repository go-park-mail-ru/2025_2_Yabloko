package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/transport"
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/repository"
	"apple_backend/store_service/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type StoreUsecaseInterface interface {
	GetStore(ctx context.Context, id string) (*domain.StoreAgg, error)
	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error)
	CreateStore(ctx context.Context, name, description, cityID, address, cardImg, openAt, closedAt string, rating float64) error
	GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error)
	GetCities(ctx context.Context) ([]*domain.City, error)
	GetTags(ctx context.Context) ([]*domain.StoreTag, error)
}

type StoreHandler struct {
	uc StoreUsecaseInterface
	rs *http_response.ResponseSender
}

func NewStoreHandler(uc StoreUsecaseInterface, log logger.Logger) *StoreHandler {
	return &StoreHandler{
		uc: uc,
		rs: http_response.NewResponseSender(log),
	}
}

func NewStoreRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog logger.Logger) {
	storeRepo := repository.NewStoreRepoPostgres(db, appLog)
	storeUC := usecase.NewStoreUsecase(storeRepo)
	storeHandler := NewStoreHandler(storeUC, appLog)

	mux.HandleFunc(apiPrefix+"stores/{id}", storeHandler.GetStore)
	mux.HandleFunc(apiPrefix+"stores", storeHandler.GetStores)
	mux.HandleFunc(apiPrefix+"stores/{id}/reviews", storeHandler.GetStoreReview)
	mux.HandleFunc(apiPrefix+"stores/cities", storeHandler.GetCities)
	mux.HandleFunc(apiPrefix+"stores/tags", storeHandler.GetTags)
}

func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "CreateStore", domain.ErrHTTPMethod, nil)
		return
	}

	req := &domain.Store{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "CreateStore", domain.ErrRequestParams, err)
		return
	}

	err := h.uc.CreateStore(r.Context(), req.Name, req.Description, req.CityID, req.Address, req.CardImg, req.OpenAt,
		req.ClosedAt, req.Rating)
	if err != nil {
		if errors.Is(err, domain.ErrStoreExist) {
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "CreateStore", domain.ErrStoreExist, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "CreateStore", domain.ErrInternalServer, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *StoreHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetStore", domain.ErrHTTPMethod, nil)
		return
	}

	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetStore", domain.ErrRequestParams, nil)
		return
	}

	store, err := h.uc.GetStore(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetStore", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetStore", domain.ErrInternalServer, err)
		return
	}

	responseStore := transport.ToStoreResponse(store)
	h.rs.Send(r.Context(), w, http.StatusOK, responseStore)
}

func (h *StoreHandler) GetStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetStores", domain.ErrHTTPMethod, nil)
		return
	}

	req := &domain.StoreFilter{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetStores", domain.ErrRequestParams, nil)
		return
	}

	stores, err := h.uc.GetStores(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetStores", domain.ErrRowsNotFound, nil)
			return
		} else if errors.Is(err, domain.ErrRequestParams) {
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetStores", domain.ErrRequestParams, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetStores", domain.ErrInternalServer, err)
		return
	}

	responseStores := transport.ToStoreResponses(stores)
	h.rs.Send(r.Context(), w, http.StatusOK, responseStores)
}

func (h *StoreHandler) GetStoreReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetStoreReview", domain.ErrHTTPMethod, nil)
		return
	}

	storeID := r.PathValue("id")
	if _, err := uuid.Parse(storeID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetStoreReview", domain.ErrRequestParams, nil)
		return
	}

	reviews, err := h.uc.GetStoreReview(r.Context(), storeID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetStoreReview", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetStoreReview", domain.ErrInternalServer, err)
		return
	}

	responseReview := transport.ToStoreReviews(reviews)
	h.rs.Send(r.Context(), w, http.StatusOK, responseReview)
}

func (h *StoreHandler) GetCities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetCities", domain.ErrHTTPMethod, nil)
		return
	}

	cities, err := h.uc.GetCities(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetCities", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetCities", domain.ErrInternalServer, err)
		return
	}

	responseCities := transport.ToCityResponses(cities)
	h.rs.Send(r.Context(), w, http.StatusOK, responseCities)
}

func (h *StoreHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetTags", domain.ErrHTTPMethod, nil)
		return
	}

	tags, err := h.uc.GetTags(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetTags", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetTags", domain.ErrInternalServer, err)
		return
	}

	responseTags := transport.ToTagResponses(tags)
	h.rs.Send(r.Context(), w, http.StatusOK, responseTags)
}
