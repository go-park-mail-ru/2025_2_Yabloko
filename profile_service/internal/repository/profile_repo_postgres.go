package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
		SELECT id, email, name, phone, city_id, address, created_at, updated_at 
		FROM account 
		WHERE id = $1
	`
	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{id},
		"operation": "GetProfile",
	})

	profile := &domain.Profile{}
	var name, phone, cityID, address *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.ID,
		&profile.Email,
		&name,
		&phone,
		&cityID,
		&address,
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

	profile.Name = name
	profile.Phone = phone
	profile.CityID = cityID
	profile.Address = address

	r.log.Debug(ctx, "GetProfile завершено успешно", map[string]interface{}{"id": id})
	return profile, nil
}

func (r *ProfileRepoPostgres) GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error) {
	r.log.Debug(ctx, "GetProfileByEmail начало обработки", map[string]interface{}{"email": email})

	query := `
		SELECT id, email, name, phone, city_id, address, created_at, updated_at 
		FROM account 
		WHERE email = $1
	`

	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{email},
		"operation": "GetProfileByEmail",
	})

	profile := &domain.Profile{}
	var name, phone, cityID, address *string

	err := r.db.QueryRow(ctx, query, email).Scan(
		&profile.ID,
		&profile.Email,
		&name,
		&phone,
		&cityID,
		&address,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn(ctx, "GetProfileByEmail профиль не найден", map[string]interface{}{"email": email})
			return nil, domain.ErrProfileNotFound
		}
		r.log.Error(ctx, "GetProfileByEmail ошибка БД", map[string]interface{}{"err": err, "email": email})
		return nil, err
	}

	profile.Name = name
	profile.Phone = phone
	profile.CityID = cityID
	profile.Address = address

	r.log.Debug(ctx, "GetProfileByEmail завершено успешно", map[string]interface{}{"email": email})
	return profile, nil
}

func (r *ProfileRepoPostgres) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	r.log.Debug(ctx, "UpdateProfile начало обработки", map[string]interface{}{"id": profile.ID})

	query := `
		UPDATE account 
		SET name = $1, phone = $2, city_id = $3, address = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`

	r.log.Debug(ctx, "SQL запрос", map[string]interface{}{
		"query":     query,
		"params":    []interface{}{profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID},
		"operation": "UpdateProfile",
	})

	err := r.db.QueryRow(ctx, query,
		profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID,
	).Scan(&profile.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn(ctx, "UpdateProfile профиль не найден", map[string]interface{}{"id": profile.ID})
			return domain.ErrProfileNotFound
		}
		r.log.Error(ctx, "UpdateProfile ошибка БД", map[string]interface{}{"err": err, "profile": profile})
		return err
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
