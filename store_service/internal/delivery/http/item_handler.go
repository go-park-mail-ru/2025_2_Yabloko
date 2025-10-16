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

func NewItemHandler(uc ItemUsecaseInterface, log *logger.Logger) *ItemHandler {
	return &ItemHandler{
		uc: uc,
		rs: http_response.NewResponseSender(log),
	}
}

func NewItemRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog *logger.Logger) {
	itemRepo := repository.NewItemRepoPostgres(db, appLog)
	itemUC := usecase.NewItemUsecase(itemRepo)
	itemHandler := NewItemHandler(itemUC, appLog)

	mux.HandleFunc(apiPrefix+"/stores/{id}/items", itemHandler.GetItems)
	mux.HandleFunc(apiPrefix+"/stores/{id}/item-types", itemHandler.GetItemTypes)
}

// GetItemTypes godoc
// @Summary      Получить список типов товара
// @Description  Возвращает список типов товара по ID магазина
// @Tags         items
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID магазина (UUID)"
// @Success      200  {array}   transport.ItemType
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параметры"
// @Failure      404  {object}   http_response.ErrResponse  "Типы товаров не найдены"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /stores/{id}/item-types [get]
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

// GetItems godoc
// @Summary      Получить список товаров
// @Description  Возвращает список товаров по ID типа товара
// @Tags         items
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID типа товара (UUID)"
// @Success      200  {array}   transport.Item
// @Failure      400  {object}  http_response.ErrResponse  "Некорректные параметры"
// @Failure      404  {object}   http_response.ErrResponse  "Товары не найдены"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /stores/{id}/items [get]
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
