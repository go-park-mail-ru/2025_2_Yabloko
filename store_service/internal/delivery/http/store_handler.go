package http

import (
	"apple_backend/custom_errors"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/middlewares"
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/repository"
	"apple_backend/store_service/internal/usecase"
	"encoding/json"
	"log/slog"
	"net/http"
)

type StoreHandler struct {
	uc *usecase.StoreUsecase
	rs *http_response.ResponseSender
}

func NewStoreHandler(uc *usecase.StoreUsecase, log *logger.Logger) *StoreHandler {
	return &StoreHandler{uc: uc, rs: http_response.NewResponseSender(log)}
}

func NewStoreRouter(db repository.PgxIface, apiPrefix string) http.Handler {
	log := logger.NewLogger("./logs/store_log.log", slog.LevelDebug)

	storeRepo := repository.NewStoreRepoPostgres(db)
	storeUC := usecase.NewStoreUsecase(storeRepo)
	storeHandler := NewStoreHandler(storeUC, log)

	mux := http.NewServeMux()

	mux.HandleFunc(apiPrefix+"/stores/{id}", storeHandler.GetStore)
	mux.HandleFunc(apiPrefix+"/stores", storeHandler.GetStores)
	//mux.HandleFunc(apiPrefix+"/stores", storeHandler.CreateStore)

	return middlewares.CorsMiddleware(middlewares.AccessLog(log, mux))
}

func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, custom_errors.HTTPMethodErr, nil)
		return
	}

	req := &domain.Store{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, custom_errors.InvalidJSONErr, err)
		return
	}

	err := h.uc.CreateStore(req.Name, req.Description, req.CityID, req.Address, req.CardImg, req.OpenAt, req.ClosedAt,
		req.Rating)
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusCreated, nil)
}

func (h *StoreHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, custom_errors.HTTPMethodErr, nil)
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, custom_errors.InvalidJSONErr, nil)
		return
	}

	store, err := h.uc.GetStore(id)
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusNotFound, custom_errors.NotExistErr, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusOK, store)
}

func (h *StoreHandler) GetStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, custom_errors.HTTPMethodErr, nil)
		return
	}

	req := &domain.StoreFilter{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, custom_errors.InvalidJSONErr, nil)
		return
	}

	stores, err := h.uc.GetStores(req)
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusOK, stores)
}
