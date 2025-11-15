package repository

import (
	"context"
	"errors"
	"log/slog"

	"apple_backend/pkg/logger"
	"apple_backend/support_service/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type RatingRepoPostgres struct {
	db PgxIface
}

func NewRatingRepoPostgres(db PgxIface) *RatingRepoPostgres {
	return &RatingRepoPostgres{db: db}
}

func (r *RatingRepoPostgres) CreateRating(ctx context.Context, rating *domain.Rating) error {
	log := logger.FromContext(ctx)

	query := `INSERT INTO support_rating (id, ticket_id, rating, comment, created_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := r.db.Exec(ctx, query, rating.ID, rating.TicketID, rating.Rating, rating.Comment, rating.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.WarnContext(ctx, "CreateRating duplicate", slog.Any("err", err))
			return domain.ErrRatingExists
		}
		log.ErrorContext(ctx, "CreateRating db error", slog.Any("err", err))
		return err
	}
	return nil
}

func (r *RatingRepoPostgres) GetRatingByTicket(ctx context.Context, ticketID string) (*domain.Rating, error) {
	log := logger.FromContext(ctx)
	query := `SELECT id, ticket_id, rating, comment, created_at FROM support_rating WHERE ticket_id=$1`
	row := r.db.QueryRow(ctx, query, ticketID)

	var rating domain.Rating
	err := row.Scan(&rating.ID, &rating.TicketID, &rating.Rating, &rating.Comment, &rating.CreatedAt)
	if err != nil {
		log.ErrorContext(ctx, "GetRatingByTicket db error", slog.Any("err", err))
		return nil, domain.ErrRowsNotFound
	}
	return &rating, nil
}

func (r *RatingRepoPostgres) GetAverageRating(ctx context.Context) (float64, error) {
	log := logger.FromContext(ctx)
	query := `SELECT AVG(rating) FROM support_rating`
	row := r.db.QueryRow(ctx, query)
	var avg float64
	err := row.Scan(&avg)
	if err != nil {
		log.ErrorContext(ctx, "GetAverageRating db error", slog.Any("err", err))
		return 0, err
	}
	return avg, nil
}
