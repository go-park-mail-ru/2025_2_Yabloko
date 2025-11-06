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
	db  PgxIface
	log logger.Logger
}

func NewProfileRepoPostgres(db PgxIface, log logger.Logger) *ProfileRepoPostgres {
	return &ProfileRepoPostgres{db: db, log: log}
}

//go:embed sql/profile/get_profile.sql
var getProfileQuery string

func (r *ProfileRepoPostgres) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {

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
			r.log.Warn("GetProfile профиль не найден", slog.String("id", id))
			return nil, domain.ErrProfileNotFound
		}
		r.log.Error("GetProfile ошибка БД", slog.Any("err", err), slog.Any("id", id))
		return nil, err
	}

	r.log.Debug("GetProfile завершено успешно", slog.String("id", id))
	return p, nil
}

//go:embed sql/profile/update_profile.sql
var updateProfileQuery string

func (r *ProfileRepoPostgres) UpdateProfile(ctx context.Context, p *domain.Profile) error {

	r.log.Debug("UpdateProfile начало обработки", slog.String("id", p.ID))

	res, err := r.db.Exec(ctx, updateProfileQuery,
		p.Name,
		p.Phone,
		p.CityID,
		p.Address,
		p.AvatarURL,
		p.ID,
	)

	if err != nil {
		r.log.Error("UpdateProfile ошибка БД", slog.String("id", p.ID), slog.Any("err", err))
		return err
	}

	if res.RowsAffected() == 0 {
		r.log.Warn("UpdateProfile профиль не найден", slog.String("id", p.ID))
		return domain.ErrProfileNotFound
	}

	r.log.Debug("UpdateProfile завершено успешно", slog.String("id", p.ID))
	return nil
}

//go:embed sql/profile/delete_profile.sql
var deleteProfileQuery string

func (r *ProfileRepoPostgres) DeleteProfile(ctx context.Context, id string) error {

	r.log.Debug("DeleteProfile начало обработки", slog.String("id", id))

	res, err := r.db.Exec(ctx, deleteProfileQuery, id)

	if err != nil {
		r.log.Error("DeleteProfile ошибка БД", slog.String("id", id))
		return err
	}

	if res.RowsAffected() == 0 {
		r.log.Warn("DeleteProfile профиль не найден", slog.String("id", id))
		return domain.ErrProfileNotFound
	}

	r.log.Debug("DeleteProfile завершено успешно", slog.String("id", id))
	return nil
}
