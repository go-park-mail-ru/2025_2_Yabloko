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

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OrderUsecaseInterface interface {
	CreateOrder(ctx context.Context, userID string) (*domain.OrderInfo, error)
	UpdateOrderStatus(ctx context.Context, orderID, userID, status string) error
	GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error)
	GetOrder(ctx context.Context, orderID, userID string) (*domain.OrderInfo, error)
}

type OrderHandler struct {
	uc        OrderUsecaseInterface
	rs        *http_response.ResponseSender
	validator *validator.Validate
}

func NewOrderHandler(uc OrderUsecaseInterface, log logger.Logger) *OrderHandler {
	return &OrderHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(log),
		validator: validator.New(),
	}
}

func NewOrderRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog logger.Logger) {
	orderRepo := repository.NewOrderRepoPostgres(db, appLog)
	orderUC := usecase.NewOrderUsecase(orderRepo)
	orderHandler := NewOrderHandler(orderUC, appLog)

	mux.HandleFunc(apiPrefix+"orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			orderHandler.GetOrdersUser(w, r)
		case http.MethodPost:
			orderHandler.CreateOrder(w, r)
		default:
			orderHandler.rs.Error(r.Context(), w,
				http.StatusMethodNotAllowed, "orders", domain.ErrHTTPMethod, nil)
		}
	})

	mux.HandleFunc(apiPrefix+"orders/{id}/status", orderHandler.UpdateOrderStatus)
	mux.HandleFunc(apiPrefix+"orders/{id}", orderHandler.GetOrder)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "CreateOrder", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "CreateOrder", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "CreateOrder", domain.ErrUnauthorized, nil)
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

func (h *OrderHandler) GetOrdersUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetOrdersUser", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetOrdersUser", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetOrdersUser", domain.ErrUnauthorized, nil)
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

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetOrder", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetOrder", domain.ErrUnauthorized, nil)
		return
	}

	order, err := h.uc.GetOrder(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetOrder", domain.ErrRowsNotFound, err)
			return
		} else if errors.Is(err, domain.ErrForbidden) {
			h.rs.Error(r.Context(), w, http.StatusForbidden, "GetOrder", domain.ErrForbidden, err)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetOrder", domain.ErrInternalServer, err)
		return
	}

	orderInfo := transport.ToOrderInfoResponse(order)
	h.rs.Send(r.Context(), w, http.StatusOK, orderInfo)
}

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

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "UpdateOrderStatus", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "UpdateOrderStatus", domain.ErrUnauthorized, nil)
		return
	}

	statusReq := &transport.OrderStatus{}
	if err := json.NewDecoder(r.Body).Decode(statusReq); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, err)
		return
	}

	if err := h.validator.Struct(statusReq); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, err)
		return
	}

	if statusReq.Status != "cancelled" {
		h.rs.Error(r.Context(), w, http.StatusForbidden, "UpdateOrderStatus", domain.ErrForbidden,
			errors.New("Отменять можно только заказы `в ожидании оплаты`"))
		return
	}

	err := h.uc.UpdateOrderStatus(r.Context(), id, userID, statusReq.Status)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "UpdateOrderStatus", domain.ErrRowsNotFound, err)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusForbidden, "UpdateOrderStatus", domain.ErrForbidden, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
