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
	"net/http"

	"github.com/google/uuid"
)

type ItemUsecaseInterface interface {
	GetItemTypes(ctx context.Context, id string) ([]*domain.ItemType, error)
	GetItems(ctx context.Context, id string) ([]*domain.ItemAgg, error)
}

type ItemHandler struct {
	uc ItemUsecaseInterface
	rs *http_response.ResponseSender
}

func NewItemHandler(uc ItemUsecaseInterface, log logger.Logger) *ItemHandler {
	return &ItemHandler{
		uc: uc,
		rs: http_response.NewResponseSender(log),
	}
}

func NewItemRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog logger.Logger) {
	itemRepo := repository.NewItemRepoPostgres(db, appLog)
	itemUC := usecase.NewItemUsecase(itemRepo)
	itemHandler := NewItemHandler(itemUC, appLog)

	mux.HandleFunc(apiPrefix+"stores/{id}/items", itemHandler.GetItems)
	mux.HandleFunc(apiPrefix+"stores/{id}/item-types", itemHandler.GetItemTypes)
}

func (h *ItemHandler) GetItemTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetItemTypes", domain.ErrHTTPMethod, nil)
		return
	}

	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetItemTypes", domain.ErrRequestParams, nil)
		return
	}

	itemTypes, err := h.uc.GetItemTypes(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetItemTypes", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetItemTypes", domain.ErrInternalServer, err)
		return
	}

	responseItemTypes := transport.ToItemTypesResponse(itemTypes)
	h.rs.Send(r.Context(), w, http.StatusOK, responseItemTypes)
}

func (h *ItemHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetItems", domain.ErrHTTPMethod, nil)
		return
	}

	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetItems", domain.ErrRequestParams, nil)
		return
	}

	items, err := h.uc.GetItems(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetItems", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetItems", domain.ErrInternalServer, err)
		return
	}

	responseItems := transport.ToItemsResponse(items)
	h.rs.Send(r.Context(), w, http.StatusOK, responseItems)
}
