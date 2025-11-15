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
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CartUsecaseInterface interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	UpdateCart(ctx context.Context, userId string, updateCart *domain.CartUpdate) error
}

type CartHandler struct {
	uc        CartUsecaseInterface
	rs        *http_response.ResponseSender
	validator *validator.Validate
}

func NewCartHandler(uc CartUsecaseInterface) *CartHandler {
	return &CartHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(logger.Global()),
		validator: validator.New(),
	}
}

func NewCartRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string) {
	cartRepo := repository.NewCartRepoPostgres(db)
	cartUC := usecase.NewCartUsecase(cartRepo)
	cartHandler := NewCartHandler(cartUC)

	mux.HandleFunc(apiPrefix+"cart", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cartHandler.GetCart(w, r)
		case http.MethodPut:
			cartHandler.UpdateCart(w, r)
		default:
			ctx := r.Context()
			log := logger.FromContext(ctx)
			log.WarnContext(ctx, "handler cart wrong method", slog.String("method", r.Method))
			cartHandler.rs.Error(ctx, w, http.StatusMethodNotAllowed, "cart", domain.ErrHTTPMethod, nil)
		}
	})
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetCart start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler GetCart wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetCart", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler GetCart unauthorized - no user ID")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetCart", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		log.WarnContext(ctx, "handler GetCart invalid user ID", slog.String("user_id", userID))
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetCart", domain.ErrUnauthorized, nil)
		return
	}

	log.DebugContext(ctx, "handler GetCart processing", slog.String("user_id", userID))

	cart, err := h.uc.GetCart(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "handler GetCart usecase failed", slog.Any("err", err), slog.String("user_id", userID))
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "GetCart", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetCart", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler GetCart success",
		slog.String("user_id", userID),
		slog.Int("items_count", len(cart.Items)))
	respCart := transport.ToCartResponse(cart)
	h.rs.Send(ctx, w, http.StatusOK, respCart)
}

func (h *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler UpdateCart start")

	if r.Method != http.MethodPut {
		log.WarnContext(ctx, "handler UpdateCart wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "UpdateCart", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler UpdateCart unauthorized - no user ID")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "UpdateCart", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		log.WarnContext(ctx, "handler UpdateCart invalid user ID", slog.String("user_id", userID))
		h.rs.Error(ctx, w, http.StatusUnauthorized, "UpdateCart", domain.ErrUnauthorized, nil)
		return
	}

	log.DebugContext(ctx, "handler UpdateCart processing", slog.String("user_id", userID))

	cartUpdate := &transport.CartUpdate{}
	if err := json.NewDecoder(r.Body).Decode(cartUpdate); err != nil {
		log.ErrorContext(ctx, "handler UpdateCart decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, err)
		return
	}

	if err := h.validator.Struct(cartUpdate); err != nil {
		log.WarnContext(ctx, "handler UpdateCart validation failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, err)
		return
	}

	log.DebugContext(ctx, "handler UpdateCart validation passed",
		slog.Int("items_count", len(cartUpdate.Items)))

	err := h.uc.UpdateCart(ctx, userID, transport.FromCartUpdate(cartUpdate))
	if err != nil {
		log.ErrorContext(ctx, "handler UpdateCart usecase failed", slog.Any("err", err), slog.String("user_id", userID))

		if errors.Is(err, domain.ErrRequestParams) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, err)
			return
		}

		if errors.Is(err, domain.ErrInvalidQuantity) {
			h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateCart", domain.ErrRequestParams, err)
			return
		}

		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "UpdateCart", domain.ErrRowsNotFound, err)
			return
		}

		h.rs.Error(ctx, w, http.StatusInternalServerError, "UpdateCart", domain.ErrInternalServer, err)
		return
	}

	log.InfoContext(ctx, "handler UpdateCart success",
		slog.String("user_id", userID),
		slog.Int("items_count", len(cartUpdate.Items)))
	w.WriteHeader(http.StatusNoContent)
}
