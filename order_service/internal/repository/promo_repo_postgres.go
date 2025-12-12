package repository

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/logger"
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:embed sql/promo/get_by_code.sql
var getPromocodeByCode string

//go:embed sql/promo/insert_promocode_account.sql
var insertPromocodeAccount string

type PromoRepoPostgres struct {
	db PgxIface
}

func NewPromoRepoPostgres(db PgxIface) *PromoRepoPostgres {
	return &PromoRepoPostgres{db: db}
}

func (r *PromoRepoPostgres) GetActiveForUserByCode(ctx context.Context, userID, code string, now time.Time) (*domain.Promocode, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetActiveForUserByCode start",
		slog.String("user_id", userID),
		slog.String("code", code),
	)

	var p domain.Promocode
	err := r.db.QueryRow(ctx, getPromocodeByCode, userID, code, now).Scan(
		&p.ID,
		&p.Code,
		&p.RelativeDiscount,
		&p.AbsoluteDiscount,
		&p.StartAt,
		&p.EndAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo GetActiveForUserByCode not found or already used",
				slog.String("user_id", userID),
				slog.String("code", code),
			)
			return nil, domain.ErrRowsNotFound
		}
		log.ErrorContext(ctx, "repo GetActiveForUserByCode query failed",
			slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	return &p, nil
}

func (r *PromoRepoPostgres) MarkUsed(ctx context.Context, userID, promoID string) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo MarkUsed start",
		slog.String("user_id", userID),
		slog.String("promocode_id", promoID),
	)

	_, err := r.db.Exec(ctx, insertPromocodeAccount, uuid.New().String(), userID, promoID)
	if err != nil {
		// unique (user_id, promocode_id) защитит от повторного использования
		log.ErrorContext(ctx, "repo MarkUsed insert failed", slog.Any("err", err))
		return domain.ErrInternalServer
	}

	return nil
}
