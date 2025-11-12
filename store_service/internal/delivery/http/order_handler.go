package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/middlewares"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"apple_backend/store_service/internal/delivery/transport"
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/repository"
	"apple_backend/store_service/internal/usecase"

	"github.com/go-playground/validator/v10"
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

func NewOrderHandler(uc OrderUsecaseInterface) *OrderHandler {
	return &OrderHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(logger.Global()),
		validator: validator.New(),
	}
}

func NewOrderRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string) {
	orderRepo := repository.NewOrderRepoPostgres(db)
	orderUC := usecase.NewOrderUsecase(orderRepo)
	orderHandler := NewOrderHandler(orderUC)

	mux.HandleFunc(apiPrefix+"orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			orderHandler.GetOrdersUser(w, r)
		case http.MethodPost:
			orderHandler.CreateOrder(w, r)
		default:
			ctx := r.Context()
			log := logger.FromContext(ctx)
			log.WarnContext(ctx, "handler orders wrong method", slog.String("method", r.Method))
			orderHandler.rs.Error(ctx, w, http.StatusMethodNotAllowed, "orders", domain.ErrHTTPMethod, nil)
		}
	})

	mux.HandleFunc(apiPrefix+"orders/{id}/status", orderHandler.UpdateOrderStatus)
	mux.HandleFunc(apiPrefix+"orders/{id}", orderHandler.GetOrder)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler CreateOrder start")

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler CreateOrder unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "CreateOrder", domain.ErrUnauthorized, nil)
		return
	}

	orderInfo, err := h.uc.CreateOrder(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "handler CreateOrder failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrCartEmpty):
			h.rs.Error(ctx, w, http.StatusBadRequest, "CreateOrder", domain.ErrRequestParams, err)
		case errors.Is(err, domain.ErrInternalServer):
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreateOrder", domain.ErrInternalServer, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreateOrder", domain.ErrInternalServer, err)
		}
		return
	}

	log.InfoContext(ctx, "handler CreateOrder success",
		slog.String("order_id", orderInfo.ID),
		slog.Int("items_count", len(orderInfo.Items)))

	order := transport.ToOrderInfoResponse(orderInfo)
	h.rs.Send(ctx, w, http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetOrdersUser start")

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler GetOrdersUser unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetOrdersUser", domain.ErrUnauthorized, nil)
		return
	}

	orders, err := h.uc.GetOrdersUser(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "handler GetOrdersUser failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrRowsNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "GetOrdersUser", domain.ErrRowsNotFound, err)
		case errors.Is(err, domain.ErrInternalServer):
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetOrdersUser", domain.ErrInternalServer, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetOrdersUser", domain.ErrInternalServer, err)
		}
		return
	}

	log.InfoContext(ctx, "handler GetOrdersUser success", slog.Int("orders_count", len(orders)))
	ordersInfo := transport.ToOrdersResponse(orders)
	h.rs.Send(ctx, w, http.StatusOK, ordersInfo)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetOrder start")

	id := r.PathValue("id")
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler GetOrder unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetOrder", domain.ErrUnauthorized, nil)
		return
	}

	order, err := h.uc.GetOrder(ctx, id, userID)
	if err != nil {
		log.ErrorContext(ctx, "handler GetOrder failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrRowsNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "GetOrder", domain.ErrRowsNotFound, err)
		case errors.Is(err, domain.ErrForbidden):
			h.rs.Error(ctx, w, http.StatusForbidden, "GetOrder", domain.ErrForbidden, err)
		case errors.Is(err, domain.ErrInternalServer):
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetOrder", domain.ErrInternalServer, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetOrder", domain.ErrInternalServer, err)
		}
		return
	}

	log.InfoContext(ctx, "handler GetOrder success", slog.String("order_id", id))
	orderInfo := transport.ToOrderInfoResponse(order)
	h.rs.Send(ctx, w, http.StatusOK, orderInfo)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler UpdateOrderStatus start")

	id := r.PathValue("id")
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler UpdateOrderStatus unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "UpdateOrderStatus", domain.ErrUnauthorized, nil)
		return
	}

	statusReq := &transport.OrderStatus{}
	if err := json.NewDecoder(r.Body).Decode(statusReq); err != nil {
		log.ErrorContext(ctx, "handler UpdateOrderStatus decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, err)
		return
	}

	if err := h.validator.Struct(statusReq); err != nil {
		log.WarnContext(ctx, "handler UpdateOrderStatus validation failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, err)
		return
	}

	if statusReq.Status != "cancelled" {
		log.WarnContext(ctx, "handler UpdateOrderStatus invalid status", slog.String("status", statusReq.Status))
		h.rs.Error(ctx, w, http.StatusForbidden, "UpdateOrderStatus", domain.ErrForbidden,
			errors.New("Отменять можно только заказы `в ожидании оплаты`"))
		return
	}

	err := h.uc.UpdateOrderStatus(ctx, id, userID, statusReq.Status)
	if err != nil {
		log.ErrorContext(ctx, "handler UpdateOrderStatus failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrRowsNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "UpdateOrderStatus", domain.ErrRowsNotFound, err)
		case errors.Is(err, domain.ErrInternalServer):
			h.rs.Error(ctx, w, http.StatusInternalServerError, "UpdateOrderStatus", domain.ErrInternalServer, err)
		default:
			h.rs.Error(ctx, w, http.StatusForbidden, "UpdateOrderStatus", domain.ErrForbidden, err)
		}
		return
	}

	log.InfoContext(ctx, "handler UpdateOrderStatus success", slog.String("order_id", id))
	w.WriteHeader(http.StatusNoContent)
}
