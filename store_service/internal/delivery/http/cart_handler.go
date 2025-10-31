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

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CartUsecaseInterface interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	UpdateCart(ctx context.Context, cartID string, updateCart *domain.CartUpdate) error
}

type CartHandler struct {
	uc        CartUsecaseInterface
	rs        *http_response.ResponseSender
	validator *validator.Validate
}

func NewCartHandler(uc CartUsecaseInterface, log *logger.Logger) *CartHandler {
	return &CartHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(log),
		validator: validator.New(),
	}
}

func NewCartRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog *logger.Logger) {
	cartRepo := repository.NewCartRepoPostgres(db, appLog)
	cartUC := usecase.NewCartUsecase(cartRepo)
	cartHandler := NewCartHandler(cartUC, appLog)

	mux.HandleFunc(apiPrefix+"/users/{id}/cart", cartHandler.GetCart)
	mux.HandleFunc(apiPrefix+"/carts/{id}", cartHandler.GetCart)
}

// GetCart godoc
// @Summary      Получить корзину по ID пользователя
// @Description  Возвращает корзину с товарами
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID пользователя (UUID)"
// @Success      200  {array}   transport.Cart
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параметры"
// @Failure      404  {object}   http_response.ErrResponse  "Пустая корзина"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /users/{id}/cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetCart", domain.ErrHTTPMethod, nil)
		return
	}

	userID := r.PathValue("id")
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetCart", domain.ErrRequestParams, nil)
		return
	}

	cart, err := h.uc.GetCart(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetCart", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetCart", domain.ErrInternalServer, err)
		return
	}

	respCart := transport.ToCartResponse(cart)
	h.rs.Send(r.Context(), w, http.StatusOK, respCart)
}

// UpdateCart godoc
// @Summary      Заменить товары в корзине
// @Description  Очищает корзину и добавляет новый набор товаров
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID корзины (UUID)"
// @Param cart_items body transport.CartUpdate true "Список товаров в корзине"
// @Success      200  {object}   transport.UpdateResponse
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параетры запроса"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /carts/{id} [put]
func (h *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "UpdateCart", domain.ErrHTTPMethod, nil)
		return
	}

	cartID := r.PathValue("id")
	if _, err := uuid.Parse(cartID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, nil)
		return
	}

	cartUpdate := &transport.CartUpdate{}
	if err := json.NewDecoder(r.Body).Decode(cartUpdate); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, err)
		return
	}

	if err := h.validator.Struct(cartUpdate); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, err)
		return
	}

	err := h.uc.UpdateCart(r.Context(), cartID, transport.FromCartUpdate(cartUpdate))
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "UpdateCart", domain.ErrInternalServer, err)
		return
	}

	response := &transport.UpdateResponse{ID: cartID}
	h.rs.Send(r.Context(), w, http.StatusOK, response)
}
