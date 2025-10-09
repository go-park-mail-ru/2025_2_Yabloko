package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/middlewares"
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
	GetStore(ctx context.Context, id string) (*domain.Store, error)
	GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.Store, error)
	CreateStore(ctx context.Context, name, description, cityID, address, cardImg, openAt, closedAt string, rating float64) error
}

type StoreHandler struct {
	uc StoreUsecaseInterface
	rs *http_response.ResponseSender
}

func NewStoreHandler(uc StoreUsecaseInterface, log *logger.Logger) *StoreHandler {
	return &StoreHandler{
		uc: uc,
		rs: http_response.NewResponseSender(log),
	}
}

func NewStoreRouter(db repository.PgxIface, apiPrefix string, appLog, accessLog *logger.Logger) http.Handler {
	storeRepo := repository.NewStoreRepoPostgres(db, appLog)
	storeUC := usecase.NewStoreUsecase(storeRepo)
	storeHandler := NewStoreHandler(storeUC, appLog)

	mux := http.NewServeMux()

	mux.HandleFunc(apiPrefix+"/stores/{id}", storeHandler.GetStore)
	mux.HandleFunc(apiPrefix+"/stores", storeHandler.GetStores)
	//mux.HandleFunc(apiPrefix+"/stores", storeHandler.CreateStore)

	return middlewares.CorsMiddleware(middlewares.AccessLog(accessLog, mux))
}

func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "CreateStore", domain.ErrHTTPMethod, nil)
		return
	}

	req := &domain.Store{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "CreateStore", domain.ErrRequestParams, err)
		return
	}

	err := h.uc.CreateStore(r.Context(), req.Name, req.Description, req.CityID, req.Address, req.CardImg, req.OpenAt,
		req.ClosedAt, req.Rating)
	if err != nil {
		if errors.Is(err, domain.ErrStoreExist) {
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "CreateStore", domain.ErrStoreExist, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "CreateStore", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusCreated, nil)
}

func (h *StoreHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetStore", domain.ErrHTTPMethod, nil)
	}

	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetStore", domain.ErrRequestParams, nil)
		return
	}

	store, err := h.uc.GetStore(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetStore", domain.ErrStoreNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusNotFound, "GetStore", domain.ErrInternalServer, err)
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetStores", domain.ErrRequestParams, nil)
		return
	}

	stores, err := h.uc.GetStores(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetStores", domain.ErrStoreNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetStores", domain.ErrStoreNotFound, err)
		return
	}

	responseStores := transport.ToStoreResponses(stores)
	h.rs.Send(r.Context(), w, http.StatusOK, responseStores)
}
