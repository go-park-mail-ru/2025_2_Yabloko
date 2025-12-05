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

	"github.com/google/uuid"
)

type ItemUsecaseInterface interface {
	GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error)
	GetItems(ctx context.Context, storeID string, itemTypes []string, sorted string, desc bool) ([]*domain.ItemAgg, error)
}

type ItemHandler struct {
	uc ItemUsecaseInterface
	rs *http_response.ResponseSender
}

func NewItemHandler(uc ItemUsecaseInterface) *ItemHandler {
	return &ItemHandler{
		uc: uc,
		rs: http_response.NewResponseSender(logger.Global()),
	}
}

func NewItemRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string) {
	itemRepo := repository.NewItemRepoPostgres(db)
	itemUC := usecase.NewItemUsecase(itemRepo)
	itemHandler := NewItemHandler(itemUC)

	mux.HandleFunc(apiPrefix+"stores/{id}/items", itemHandler.GetItems)
	mux.HandleFunc(apiPrefix+"stores/{id}/item-types", itemHandler.GetItemTypes)
}

func (h *ItemHandler) GetItemTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetItemTypes start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetItemTypes wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetItemTypes", domain.ErrHTTPMethod, nil)
		return
	}

	storeID := r.PathValue("id")
	if _, err := uuid.Parse(storeID); err != nil {
		log.WarnContext(ctx, "handler GetItemTypes invalid id", slog.String("id", storeID))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetItemTypes", domain.ErrRequestParams, nil)
		return
	}

	itemTypes, err := h.uc.GetItemTypes(ctx, storeID)
	if err != nil {
		log.ErrorContext(ctx, "handler GetItemTypes usecase failed", slog.Any("err", err), slog.String("store_id", storeID))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetItemTypes", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetItemTypes", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetItemTypes success",
		slog.String("store_id", storeID),
		slog.Int("types_count", len(itemTypes)))
	responseItemTypes := transport.ToItemTypeResponses(itemTypes)
	h.rs.Send(ctx, w, http.StatusOK, responseItemTypes)
}

func (h *ItemHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetItems start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetItems wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetItems", domain.ErrHTTPMethod, nil)
		return
	}

	storeID := r.PathValue("id")
	if _, err := uuid.Parse(storeID); err != nil {
		log.WarnContext(ctx, "handler GetItems invalid id", slog.String("id", storeID))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetItems", domain.ErrRequestParams, nil)
		return
	}

	q := r.URL.Query()

	// item_type как массив: ?item_type=...&item_type=...
	itemTypes := q["item_type"]

	sorted := q.Get("sorted")
	desc := false
	if q.Has("desc") {
		// ?desc=true
		b, err := strconv.ParseBool(q.Get("desc"))
		if err == nil {
			desc = b
		}
	}

	items, err := h.uc.GetItems(ctx, storeID, itemTypes, sorted, desc)
	if err != nil {
		log.ErrorContext(ctx, "handler GetItems usecase failed", slog.Any("err", err), slog.String("store_id", storeID))
		if errors.Is(err, domain.ErrRequestParams) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "GetItems", domain.ErrRequestParams, nil)
			return
		}
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetItems", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetItems", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetItems success",
		slog.String("store_id", storeID),
		slog.Int("items_count", len(items)))
	responseItems := transport.ToItemAggResponses(items)
	h.rs.Send(ctx, w, http.StatusOK, responseItems)
}
