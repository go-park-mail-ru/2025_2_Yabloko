package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/metrics"
	"apple_backend/pkg/middlewares"
	"apple_backend/recommendation_service/internal/delivery/transport"
	"apple_backend/recommendation_service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

type RecommendationUsecaseInterface interface {
	GetHomeRecommendations(ctx context.Context, filter *domain.HomeRecommendFilter) ([]*domain.RecommendedItem, error)
}

type RecommendationHandler struct {
	uc        RecommendationUsecaseInterface
	rs        *http_response.ResponseSender
	validator *validator.Validate
}

func NewRecommendationHandler(uc RecommendationUsecaseInterface) *RecommendationHandler {
	return &RecommendationHandler{
		uc:        uc,
		rs:        http_response.NewResponseSender(logger.Global()),
		validator: validator.New(),
	}
}

func NewRecommendationRouter(mux *http.ServeMux, h *RecommendationHandler, apiPrefix string) {
	mux.HandleFunc(apiPrefix+"recommend/home", h.GetHomeRecommendations)
}

func (h *RecommendationHandler) GetHomeRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetHomeRecommendations start")

	start := time.Now()
	mode := "unknown"

	defer func() {
		metrics.RecommendHomeDuration.WithLabelValues(mode).Observe(time.Since(start).Seconds())
		metrics.RecommendHomeTotal.WithLabelValues(mode).Inc()
	}()

	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		log.WarnContext(ctx, "handler GetHomeRecommendations unauthorized")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetHomeRecommendations", domain.ErrRequestParams, errors.New("unauthorized"))
		mode = "empty"
		return
	}

	q := r.URL.Query()
	limitStr := q.Get("limit")
	if limitStr == "" {
		limitStr = "5"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 20 {
		log.WarnContext(ctx, "handler GetHomeRecommendations invalid limit", slog.String("limit", limitStr))
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetHomeRecommendations", domain.ErrRequestParams, errors.New("invalid limit"))
		mode = "empty"
		return
	}

	filter := &domain.HomeRecommendFilter{
		UserID: userID,
		Limit:  limit,
	}

	items, err := h.uc.GetHomeRecommendations(ctx, filter)
	fmt.Println(err)
	if err != nil {
		log.ErrorContext(ctx, "handler GetHomeRecommendations failed", slog.Any("err", err))

		switch {
		case errors.Is(err, domain.ErrRequestParams):
			h.rs.Error(ctx, w, http.StatusBadRequest, "GetHomeRecommendations", domain.ErrRequestParams, err)
		case errors.Is(err, domain.ErrInternalServer):
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetHomeRecommendations", domain.ErrInternalServer, err)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetHomeRecommendations", domain.ErrInternalServer, err)
		}
		mode = "empty"
		return
	}

	if len(items) == 0 {
		mode = "empty"
	} else {
		allZero := true
		for _, it := range items {
			if it.Score > 0 {
				allZero = false
				break
			}
		}
		if allZero {
			mode = "random"
		} else {
			mode = "profile"
		}
	}

	log.InfoContext(ctx, "handler GetHomeRecommendations success", slog.Int("items_count", len(items)))
	resp := transport.ToHomeRecommendationResponse(items)
	h.rs.Send(ctx, w, http.StatusOK, resp)
}
