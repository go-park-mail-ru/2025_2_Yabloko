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

type OrderUsecaseInterface interface {
	CreateOrder(ctx context.Context, userID string) (*domain.OrderInfo, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
	GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.OrderInfo, error)
}

type OrderHandler struct {
	uc OrderUsecaseInterface
	rs *http_response.ResponseSender
}

func NewOrderHandler(uc OrderUsecaseInterface, log *logger.Logger) *OrderHandler {
	return &OrderHandler{
		uc: uc,
		rs: http_response.NewResponseSender(log),
	}
}

func NewOrderRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog *logger.Logger) {
	orderRepo := repository.NewOrderRepoPostgres(db, appLog)
	orderUC := usecase.NewOrderUsecase(orderRepo)
	orderHandler := NewOrderHandler(orderUC, appLog)

	mux.HandleFunc(apiPrefix+"/users/{id}/order", orderHandler.CreateOrder)
	mux.HandleFunc(apiPrefix+"/users/{id}/orders", orderHandler.GetOrdersUser)

	mux.HandleFunc(apiPrefix+"/orders/{id}/status", orderHandler.UpdateOrderStatus)
	mux.HandleFunc(apiPrefix+"/orders/{id}", orderHandler.GetOrder)
}

// CreateOrder godoc
// @Summary      Создать заказ пользователя из текущей корзины
// @Description  Создает заказ и возвращает информацию о заказе
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID пользователя (UUID)"
// @Success      200  {array}   transport.OrderInfo
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параметры"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /users/{id}/order [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "CreateOrder", domain.ErrHTTPMethod, nil)
		return
	}
	// TODO проверка пустоты корзины
	userID := r.PathValue("id")
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "CreateOrder", domain.ErrRequestParams, nil)
		return
	}

	orderInfo, err := h.uc.CreateOrder(r.Context(), userID)
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "CreateOrder", domain.ErrInternalServer, err)
		return
	}

	order := transport.ToOrderInfoResponse(orderInfo)
	h.rs.Send(r.Context(), w, http.StatusOK, order)
}

// GetOrdersUser godoc
// @Summary      Получить все заказвы пользователя
// @Description  Краткая информация по заказам пользователя
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID пользователя (UUID)"
// @Success      200  {object}   transport.Orders
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параетры запроса"
// @Failure      404  {object}   http_response.ErrResponse  "Отсутствуют заказы"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /users/{id}/orders [get]
func (h *OrderHandler) GetOrdersUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetOrdersUser", domain.ErrHTTPMethod, nil)
		return
	}

	userID := r.PathValue("id")
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetOrdersUser", domain.ErrRequestParams, nil)
		return
	}

	orders, err := h.uc.GetOrdersUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetOrdersUser", domain.ErrRowsNotFound, err)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetOrdersUser", domain.ErrInternalServer, err)
		return
	}

	ordersInfo := transport.ToOrdersResponse(orders)
	h.rs.Send(r.Context(), w, http.StatusOK, ordersInfo)
}

// GetOrder godoc
// @Summary      Получить информацию по заказу
// @Description  Подробная информация о заказе
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID заказа (UUID)"
// @Param cart_items body transport.CartUpdate true "Список товаров в корзине"
// @Success      200  {object}   transport.OrderInfo
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параетры запроса"
// @Failure      404  {object}   http_response.ErrResponse  "Отсутствует заказ"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetOrder", domain.ErrHTTPMethod, nil)
		return
	}

	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetOrder", domain.ErrRequestParams, nil)
		return
	}

	order, err := h.uc.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetOrder", domain.ErrRowsNotFound, err)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetOrder", domain.ErrInternalServer, err)
		return
	}

	orderInfo := transport.ToOrderInfoResponse(order)
	h.rs.Send(r.Context(), w, http.StatusOK, orderInfo)
}

// UpdateOrderStatus godoc
// @Summary      Заменить статус заказа
// @Description  Меняет статус заказа на один из возможных ('pending', 'paid', 'delivered', 'cancelled', 'on the way')
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID заказа (UUID)"
// @Param cart_items body transport.OrderStatus true "Новый статус заказа"
// @Success      200  {object}   transport.OrderInfo
// @Failure      400  {object}   http_response.ErrResponse  "Некорректные параетры запроса"
// @Failure      404  {object}   http_response.ErrResponse  "Отсутствует заказ"
// @Failure      405  {object}   http_response.ErrResponse  "Метод не поддерживается"
// @Failure      500  {object}   http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "UpdateOrderStatus", domain.ErrHTTPMethod, nil)
		return
	}

	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, nil)
		return
	}

	status := &transport.OrderStatus{}
	if err := json.NewDecoder(r.Body).Decode(status); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, err)
		return
	}

	err := h.uc.UpdateOrderStatus(r.Context(), id, status.Status)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "UpdateOrderStatus", domain.ErrRowsNotFound, err)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "UpdateOrderStatus", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusOK, nil)
}
