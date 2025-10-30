package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProfileRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewProfileRepoPostgres(db PgxIface, log *logger.Logger) *ProfileRepoPostgres {
	return &ProfileRepoPostgres{
		db:  db,
		log: log,
	}
}

func (r *ProfileRepoPostgres) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	r.log.Debug(ctx, "GetProfile начало обработки", map[string]interface{}{"id": id})

	query := `
		SELECT id, email, name, phone, city_id, address, avatar_url, created_at, updated_at 
		FROM account 
		WHERE id = $1
	`

	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{id},
		"operation": "GetProfile",
	})

	profile := &domain.Profile{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.ID,
		&profile.Email,
		&profile.Name,
		&profile.Phone,
		&profile.CityID,
		&profile.Address,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
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
	return profile, nil
}

func (r *ProfileRepoPostgres) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	r.log.Debug(ctx, "UpdateProfile начало обработки", map[string]interface{}{"id": profile.ID})

	query := `
        UPDATE account
        SET
            name       = $1,
            phone      = $2,
            city_id    = $3,
            address    = $4,
            avatar_url = $5
        WHERE id = $6
    `

	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID},
		"operation": "UpdateProfile",
	})

	cmdTag, err := r.db.Exec(ctx, query,
		profile.Name,
		profile.Phone,
		profile.CityID,
		profile.Address,
		profile.AvatarURL,
		profile.ID,
	)
	if err != nil {
		r.log.Error(ctx, "UpdateProfile ошибка БД", map[string]interface{}{"err": err, "id": profile.ID})
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		r.log.Warn(ctx, "UpdateProfile профиль не найден", map[string]interface{}{"id": profile.ID})
		return domain.ErrProfileNotFound
	}

	r.log.Debug(ctx, "UpdateProfile завершено успешно", map[string]interface{}{"id": profile.ID})
	return nil
}

func (r *ProfileRepoPostgres) DeleteProfile(ctx context.Context, id string) error {
	r.log.Debug(ctx, "DeleteProfile начало обработки", map[string]interface{}{"id": id})

	query := `DELETE FROM account WHERE id = $1`

	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{id},
		"operation": "DeleteProfile",
	})

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "DeleteProfile ошибка БД", map[string]interface{}{"err": err, "id": id})
		return err
	}

	if result.RowsAffected() == 0 {
		r.log.Warn(ctx, "DeleteProfile профиль не найден", map[string]interface{}{"id": id})
		return domain.ErrProfileNotFound
	}

	r.log.Debug(ctx, "DeleteProfile завершено успешно", map[string]interface{}{"id": id})
	return nil
}

func (r *ProfileRepoPostgres) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	r.log.Debug(ctx, "CreateProfile начало обработки", map[string]interface{}{"email": profile.Email})

	query := `
        INSERT INTO account (id, email, password_hash)
        VALUES ($1, $2, $3)
    `

	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{profile.ID, profile.Email, profile.PasswordHash},
		"operation": "CreateProfile",
	})

	_, err := r.db.Exec(ctx, query, profile.ID, profile.Email, profile.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.log.Warn(ctx, "CreateProfile конфликт email (unique)", map[string]interface{}{"email": profile.Email, "pgcode": pgErr.Code})
			return domain.ErrProfileExist
		}
		r.log.Error(ctx, "CreateProfile ошибка БД", map[string]interface{}{"err": err, "email": profile.Email})
		return err
	}

	r.log.Debug(ctx, "CreateProfile завершено успешно", map[string]interface{}{"id": profile.ID})
	return nil
}
