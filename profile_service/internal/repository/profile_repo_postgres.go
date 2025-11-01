package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	_ "embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProfileRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewProfileRepoPostgres(db PgxIface, log *logger.Logger) *ProfileRepoPostgres {
	return &ProfileRepoPostgres{db: db, log: log}
}

//go:embed sql/profile/get_profile.sql
var getProfileQuery string

func (r *ProfileRepoPostgres) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	r.log.Debug(ctx, "GetProfile начало обработки", map[string]interface{}{"id": id})

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
			r.log.Warn(ctx, "GetProfile профиль не найден", map[string]interface{}{"id": id})
			return nil, domain.ErrProfileNotFound
		}
		r.log.Error(ctx, "GetProfile ошибка БД", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	r.log.Debug(ctx, "GetProfile завершено успешно", map[string]interface{}{"id": id})
	return p, nil
}

//go:embed sql/profile/update_profile.sql
var updateProfileQuery string

func (r *ProfileRepoPostgres) UpdateProfile(ctx context.Context, p *domain.Profile) error {
	r.log.Debug(ctx, "UpdateProfile начало обработки", map[string]interface{}{"id": p.ID})

	res, err := r.db.Exec(ctx, updateProfileQuery,
		p.Name,
		p.Phone,
		p.CityID,
		p.Address,
		p.AvatarURL,
		p.ID,
	)
	if err != nil {
		r.log.Error(ctx, "UpdateProfile ошибка БД", map[string]interface{}{"err": err, "id": p.ID})
		return err
	}
	if res.RowsAffected() == 0 {
		r.log.Warn(ctx, "UpdateProfile профиль не найден", map[string]interface{}{"id": p.ID})
		return domain.ErrProfileNotFound
	}

	r.log.Debug(ctx, "UpdateProfile завершено успешно", map[string]interface{}{"id": p.ID})
	return nil
}

//go:embed sql/profile/delete_profile.sql
var deleteProfileQuery string

func (r *ProfileRepoPostgres) DeleteProfile(ctx context.Context, id string) error {
	r.log.Debug(ctx, "DeleteProfile начало обработки", map[string]interface{}{"id": id})

	res, err := r.db.Exec(ctx, deleteProfileQuery, id)
	if err != nil {
		r.log.Error(ctx, "DeleteProfile ошибка БД", map[string]interface{}{"err": err, "id": id})
		return err
	}
	if res.RowsAffected() == 0 {
		r.log.Warn(ctx, "DeleteProfile профиль не найден", map[string]interface{}{"id": id})
		return domain.ErrProfileNotFound
	}

	r.log.Debug(ctx, "DeleteProfile завершено успешно", map[string]interface{}{"id": id})
	return nil
}

//go:embed sql/profile/create_profile.sql
var createProfileQuery string

func (r *ProfileRepoPostgres) CreateProfile(ctx context.Context, p *domain.Profile) error {
	r.log.Debug(ctx, "CreateProfile начало обработки", map[string]interface{}{"email": p.Email})

	_, err := r.db.Exec(ctx, createProfileQuery, p.ID, p.Email, p.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.log.Warn(ctx, "CreateProfile конфликт email (unique)", map[string]interface{}{"email": p.Email, "pgcode": pgErr.Code})
			return domain.ErrProfileExist
		}
		r.log.Error(ctx, "CreateProfile ошибка БД", map[string]interface{}{"err": err, "email": p.Email})
		return err
	}

	r.log.Debug(ctx, "CreateProfile завершено успешно", map[string]interface{}{"id": p.ID})
	return nil
}
