package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StoreRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewStoreRepoPostgres(db PgxIface, log *logger.Logger) *StoreRepoPostgres {
	return &StoreRepoPostgres{
		db:  db,
		log: log,
	}
}

func generateQuery(filter *domain.StoreFilter) (string, []any) {
	query := `
		select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
		from store
	`
	where := []string{}
	args := []any{}

	// фильтрация по тегу
	if filter.Tag != "" {
		where = append(where, fmt.Sprintf("tag_id = $%d", len(args)+1))
		args = append(args, filter.Tag)
	}
	// если не первый запрос
	if filter.LastID != "" {
		where = append(where, fmt.Sprintf("id > $%d", len(args)+1))
		args = append(args, filter.LastID)
	}

	if len(where) > 0 {
		query += " where " + strings.Join(where, " and ")
	}

	// сортировка
	orderBy := " order by id"
	if filter.Sorted != "" {
		dir := "asc"
		if filter.Desc {
			dir = "desc"
		}
		orderBy = fmt.Sprintf(" order by $%d %s, id", len(args)+1, dir)
		args = append(args, filter.Sorted)
	}
	query += orderBy

	query += fmt.Sprintf(" limit $%d", len(args)+1)
	args = append(args, filter.Limit)
	return query, args
}

func (r *StoreRepoPostgres) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.Store, error) {
	r.log.Debug(ctx, "GetStores начало обработки", map[string]interface{}{})
	query, args := generateQuery(filter)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error(ctx, "GetStores ошибка бд", map[string]interface{}{"err": err, "filter": filter})
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		err = rows.Scan(&store.ID, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt)
		if err != nil {
			r.log.Error(ctx, "GetStores ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetStores ошибка после чтения строк",
			map[string]interface{}{"err": err, "filter": filter})
		return nil, err
	}

	if len(stores) == 0 {
		r.log.Debug(ctx, "GetStores пустой ответ", map[string]interface{}{"filter": filter})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetStores завершено успешно", map[string]interface{}{})
	return stores, nil
}

func (r *StoreRepoPostgres) GetStore(ctx context.Context, id string) (*domain.Store, error) {
	query := `
		select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
		from store
		where id = $1
	`
	r.log.Debug(ctx, "GetStore начало обработки", map[string]interface{}{})

	store := &domain.Store{}
	err := r.db.QueryRow(ctx, query, id).Scan(&store.ID, &store.Name, &store.Description,
		&store.CityID, &store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn(ctx, "GetStore пустой ответ бд", map[string]interface{}{"err": err, "id": id})
			return nil, domain.ErrRowsNotFound
		}
		r.log.Error(ctx, "GetStore ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	r.log.Debug(ctx, "GetStore завершено успешно", map[string]interface{}{})
	return store, nil
}

// CreateStore не используется
func (r *StoreRepoPostgres) CreateStore(ctx context.Context, store *domain.Store) error {
	addStore := `
		insert into store (id, name, description, city_id, address, card_img, rating, open_at, closed_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	r.log.Debug(ctx, "CreateStore начало обработки", map[string]interface{}{})

	store.ID = uuid.New().String()
	_, err := r.db.Exec(ctx, addStore, store.ID, store.Name, store.Description,
		store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			r.log.Warn(ctx, "CreateStore unique ограничение", map[string]interface{}{"err": err})
			return domain.ErrStoreExist
		}
		r.log.Error(ctx, "CreateStore ошибка бд", map[string]interface{}{"err": err, "store": store})
		return err
	}

	r.log.Debug(ctx, "CreateStore завершено успешно", map[string]interface{}{})
	return nil
}
