package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	_ "embed"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type ProfileRepoPostgres struct {
	db PgxIface
}

func NewProfileRepoPostgres(db PgxIface) *ProfileRepoPostgres {
	return &ProfileRepoPostgres{db: db}
}

//go:embed sql/profile/get_profile.sql
var getProfileQuery string

func (r *ProfileRepoPostgres) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "repo GetProfile start", slog.String("id", id))

	p := &domain.Profile{}
	err := r.db.QueryRow(ctx, getProfileQuery, id).Scan(
		&p.ID,
		&p.Email,
		&p.Name,
		&p.Phone,
		&p.CityID,
		&p.Address,
		&p.AvatarURL,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo GetProfile profile not found", slog.String("id", id))
			return nil, domain.ErrProfileNotFound
		}
		log.ErrorContext(ctx, "repo GetProfile db error", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}

	log.InfoContext(ctx, "repo GetProfile success", slog.String("id", id))
	return p, nil
}

//go:embed sql/profile/update_profile.sql
var updateProfileQuery string

func (r *ProfileRepoPostgres) UpdateProfile(ctx context.Context, p *domain.Profile) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "repo UpdateProfile start", slog.String("id", p.ID))

	res, err := r.db.Exec(ctx, updateProfileQuery,
		p.Name,
		p.Phone,
		p.CityID,
		p.Address,
		p.AvatarURL,
		p.ID,
	)

	if err != nil {
		log.ErrorContext(ctx, "repo UpdateProfile db error", slog.String("id", p.ID), slog.Any("err", err))
		return err
	}

	if res.RowsAffected() == 0 {
		log.WarnContext(ctx, "repo UpdateProfile profile not found", slog.String("id", p.ID))
		return domain.ErrProfileNotFound
	}

	log.InfoContext(ctx, "repo UpdateProfile success", slog.String("id", p.ID))
	return nil
}

//go:embed sql/profile/delete_profile.sql
var deleteProfileQuery string

func (r *ProfileRepoPostgres) DeleteProfile(ctx context.Context, id string) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "repo DeleteProfile start", slog.String("id", id))

	res, err := r.db.Exec(ctx, deleteProfileQuery, id)

	if err != nil {
		log.ErrorContext(ctx, "repo DeleteProfile db error", slog.String("id", id), slog.Any("err", err))
		return err
	}

	if res.RowsAffected() == 0 {
		log.WarnContext(ctx, "repo DeleteProfile profile not found", slog.String("id", id))
		return domain.ErrProfileNotFound
	}

	log.InfoContext(ctx, "repo DeleteProfile success", slog.String("id", id))
	return nil
}
