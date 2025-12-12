package http

import (
	"apple_backend/order_service/internal/delivery/transport"
	"apple_backend/order_service/internal/domain"
	"apple_backend/order_service/internal/repository"
	"apple_backend/order_service/internal/usecase"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/metrics"
	"apple_backend/pkg/middlewares"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OrderUsecaseInterface interface {
	CreateOrder(ctx context.Context, userID string, isFast bool, comment string, promo string) (*domain.OrderInfo, error)
	UpdateOrderStatus(ctx context.Context, orderID, userID, status string) error
	GetOrdersUser(ctx context.Context, filter *domain.OrderFilter) ([]*domain.Order, error)
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
	promoRepo := repository.NewPromoRepoPostgres(db)
	orderUC := usecase.NewOrderUsecase(orderRepo, promoRepo)
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

	var req transport.OrderCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "handler CreateOrder decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateOrder", domain.ErrRequestParams, err)
		return
	}

	orderInfo, err := h.uc.CreateOrder(ctx, userID, req.IsFast, req.Comment, req.Promo)
	if err != nil {
		log.ErrorContext(ctx, "handler CreateOrder failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrCartEmpty):
			h.rs.Error(ctx, w, http.StatusBadRequest, "CreateOrder", domain.ErrRequestParams, err)
		case errors.Is(err, domain.ErrRequestParams):
			h.rs.Error(ctx, w, http.StatusBadRequest, "CreateOrder", domain.ErrRequestParams, err)
		case errors.Is(err, domain.ErrInternalServer):
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreateOrder", domain.ErrInternalServer, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreateOrder", domain.ErrInternalServer, err)
		}
		return
	}

	totalItems := 0
	for _, store := range orderInfo.Stores {
		totalItems += len(store.Items)
	}
	log.InfoContext(ctx, "handler CreateOrder success",
		slog.String("order_id", orderInfo.ID),
		slog.Int("stores_count", len(orderInfo.Stores)),
		slog.Int("total_items", totalItems),
		slog.Bool("is_fast", orderInfo.IsFast),
		slog.String("comment", orderInfo.Comment),
	)

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

	q := r.URL.Query()
	limitStr := q.Get("limit")
	if limitStr == "" {
		log.WarnContext(ctx, "handler GetOrdersUser missing limit")
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetOrdersUser", domain.ErrRequestParams, errors.New("limit required"))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		log.WarnContext(ctx, "handler GetOrdersUser invalid limit", slog.String("limit", limitStr))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetOrdersUser", domain.ErrRequestParams, errors.New("invalid limit"))
		return
	}

	lastID := q.Get("last_id")
	if lastID != "" {
		if err := uuid.Validate(lastID); err != nil {
			log.WarnContext(ctx, "handler GetOrdersUser invalid last_id UUID", slog.String("last_id", lastID), slog.Any("err", err))
			h.rs.Error(ctx, w, http.StatusBadRequest, "GetOrdersUser", domain.ErrRequestParams, errors.New("invalid last_id UUID format"))
			return
		}
	}

	filter := &domain.OrderFilter{
		UserID: userID,
		Limit:  limit,
		LastID: lastID,
		Status: q.Get("status"),
		Desc:   q.Has("desc") && q.Get("desc") == "true",
	}

	orders, err := h.uc.GetOrdersUser(ctx, filter)
	if err != nil {
		log.ErrorContext(ctx, "handler GetOrdersUser failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrRequestParams):
			h.rs.Error(ctx, w, http.StatusBadRequest, "GetOrdersUser", domain.ErrRequestParams, err)
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

	start := time.Now()
	var order *domain.OrderInfo
	var storesCountStr string

	defer func() {
		if order != nil {
			storesCountStr = fmt.Sprintf("%d", len(order.Stores))
		} else {
			storesCountStr = "0"
		}
		metrics.OrdersGetDuration.WithLabelValues(storesCountStr).Observe(time.Since(start).Seconds())
	}()

	id := r.PathValue("id")

	if err := uuid.Validate(id); err != nil {
		log.WarnContext(ctx, "handler GetOrder invalid order_id UUID", slog.String("order_id", id), slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetOrder", domain.ErrRequestParams, errors.New("invalid order_id UUID format"))
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler GetOrder unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetOrder", domain.ErrUnauthorized, nil)
		return
	}

	var err error
	order, err = h.uc.GetOrder(ctx, id, userID)
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

	storesCountStr = fmt.Sprintf("%d", len(order.Stores))
	metrics.OrdersGetTotal.WithLabelValues(storesCountStr).Inc()

	log.InfoContext(ctx, "handler GetOrder success",
		slog.String("order_id", id),
		slog.Int("stores_count", len(order.Stores)),
		slog.Int("total_items", func() int {
			total := 0
			for _, store := range order.Stores {
				total += len(store.Items)
			}
			return total
		}()),
		slog.Bool("is_fast", order.IsFast),
		slog.String("comment", order.Comment),
	)

	orderInfo := transport.ToOrderInfoResponse(order)
	h.rs.Send(ctx, w, http.StatusOK, orderInfo)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler UpdateOrderStatus start")

	id := r.PathValue("id")

	if err := uuid.Validate(id); err != nil {
		log.WarnContext(ctx, "handler UpdateOrderStatus invalid order_id UUID", slog.String("order_id", id), slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateOrderStatus", domain.ErrRequestParams, errors.New("invalid order_id UUID format"))
		return
	}

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
