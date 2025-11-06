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

type CartUsecaseInterface interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	UpdateCart(ctx context.Context, userId string, updateCart *domain.CartUpdate) error
}

type CartHandler struct {
	uc        CartUsecaseInterface
	rs        *http_response.ResponseSender
	validator *validator.Validate
}

func NewCartHandler(uc CartUsecaseInterface, log logger.Logger) *CartHandler {
	return &CartHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(log),
		validator: validator.New(),
	}
}

func NewCartRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog logger.Logger) {
	cartRepo := repository.NewCartRepoPostgres(db, appLog)
	cartUC := usecase.NewCartUsecase(cartRepo)
	cartHandler := NewCartHandler(cartUC, appLog)

	mux.HandleFunc(apiPrefix+"cart", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cartHandler.GetCart(w, r)
		case http.MethodPut:
			cartHandler.UpdateCart(w, r)
		default:
			cartHandler.rs.Error(r.Context(), w,
				http.StatusMethodNotAllowed, "cart", domain.ErrHTTPMethod, nil)
		}
	})
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetCart", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetCart", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetCart", domain.ErrUnauthorized, nil)
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

func (h *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "UpdateCart", domain.ErrHTTPMethod, nil)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "UpdateCart", domain.ErrUnauthorized, nil)
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "UpdateCart", domain.ErrUnauthorized, nil)
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

	err := h.uc.UpdateCart(r.Context(), userID, transport.FromCartUpdate(cartUpdate))
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "UpdateCart", domain.ErrInternalServer, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
