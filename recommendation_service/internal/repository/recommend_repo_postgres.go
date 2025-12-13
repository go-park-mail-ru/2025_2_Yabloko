package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/recommendation_service/internal/domain"
	"context"
	_ "embed"
	"log/slog"
)

//go:embed sql/recommend/home_items.sql
var getHomeItems string

type RecommendationRepoPostgres struct {
	db PgxIface
}

func NewRecommendationRepoPostgres(db PgxIface) *RecommendationRepoPostgres {
	return &RecommendationRepoPostgres{
		db: db,
	}
}

func (r *RecommendationRepoPostgres) GetHomeRecommendations(
	ctx context.Context,
	filter *domain.HomeRecommendFilter,
) ([]*domain.RecommendedItem, error) {
	log := logger.FromContext(ctx)

	log.DebugContext(ctx, "repo GetHomeRecommendations start",
		slog.String("user_id", filter.UserID),
		slog.Int("limit", filter.Limit),
	)

	rows, err := r.db.Query(ctx, getHomeItems,
		filter.UserID,
		filter.Limit,
	)
	if err != nil {
		log.ErrorContext(ctx, "repo GetHomeRecommendations query failed",
			slog.Any("err", err),
			slog.String("user_id", filter.UserID),
			slog.Int("limit", filter.Limit),
		)
		return nil, domain.ErrInternalServer
	}
	defer rows.Close()

	var items []*domain.RecommendedItem

	for rows.Next() {
		var it domain.RecommendedItem

		if err := rows.Scan(
			&it.ID,
			&it.StoreID,
			&it.Name,
			&it.Price,
			&it.CardImg,
			&it.Score,
		); err != nil {
			log.ErrorContext(ctx, "repo GetHomeRecommendations scan failed",
				slog.Any("err", err),
				slog.String("user_id", filter.UserID),
				slog.Int("limit", filter.Limit),
			)
			return nil, domain.ErrInternalServer
		}

		items = append(items, &it)
	}

	if err := rows.Err(); err != nil {
		log.ErrorContext(ctx, "repo GetHomeRecommendations rows error",
			slog.Any("err", err),
			slog.String("user_id", filter.UserID),
			slog.Int("limit", filter.Limit),
		)
		return nil, domain.ErrInternalServer
	}

	if len(items) == 0 {
		log.DebugContext(ctx, "repo GetHomeRecommendations no items found",
			slog.String("user_id", filter.UserID),
			slog.Int("limit", filter.Limit),
		)
		return []*domain.RecommendedItem{}, nil
	}

	log.DebugContext(ctx, "repo GetHomeRecommendations success",
		slog.String("user_id", filter.UserID),
		slog.Int("limit", filter.Limit),
		slog.Int("items_count", len(items)),
	)

	return items, nil
}
