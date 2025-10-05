package repository

import (
	"apple_backend/custom_errors"
	"apple_backend/store_service/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StoreRepoPostgres struct {
	db PgxIface
}

func NewStoreRepoPostgres(db PgxIface) *StoreRepoPostgres {
	return &StoreRepoPostgres{db: db}
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

func (r *StoreRepoPostgres) GetStores(filter *domain.StoreFilter) ([]*domain.Store, error) {
	query, args := generateQuery(filter)

	//logger.Debug(log.LogInfo{Info: "get store", Meta: filter})
	ctxt := context.Background()

	rows, err := r.db.Query(ctxt, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			//logger.Warn(log.LogInfo{Err: custom_errors.NotExistErr, Meta: filter})
			return nil, custom_errors.NotExistErr
		}
		//logger.Error(log.LogInfo{Err: err, Meta: filter})
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		err = rows.Scan(&store.ID, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt)
		if err != nil {
			//logger.Error(log.LogInfo{Info: "get store частично завершено с ошибкой", Err: err, Meta: filter})
			return stores, err
		}
		stores = append(stores, &store)
	}
	if err = rows.Err(); err != nil {
		//logger.Error(log.LogInfo{Info: "get store завершено с ошибкой", Err: err, Meta: filter})
		return stores, err
	}

	//logger.Debug(log.LogInfo{Info: "get store завершено", Meta: filter})
	return stores, nil
}

func (r *StoreRepoPostgres) GetStore(id string) (*domain.Store, error) {
	query := `
		select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
		from store
		where id = $1
	`

	//logger.Debug(log.LogInfo{Info: "get store", Meta: filter})
	ctxt := context.Background()

	store := &domain.Store{}
	err := r.db.QueryRow(ctxt, query, id).Scan(&store.ID, &store.Name, &store.Description,
		&store.CityID, &store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			//logger.Warn(log.LogInfo{Err: custom_errors.NotExistErr, Meta: filter})
			return nil, custom_errors.NotExistErr
		}
		//logger.Error(log.LogInfo{Err: err, Meta: filter})
		return nil, err
	}

	//logger.Debug(log.LogInfo{Info: "get store завершено", Meta: filter})
	return store, nil
}

func (r *StoreRepoPostgres) CreateStore(store *domain.Store) error {
	addStore := `
		insert into store (id, name, description, city_id, address, card_img, rating, open_at, closed_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	//logger.Debug(log.LogInfo{Info: "create store", Meta: store})
	ctxt := context.Background()

	store.ID = uuid.New().String()
	_, err := r.db.Exec(ctxt, addStore, store.ID, store.Name, store.Description,
		store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			//logger.Warn(log.LogInfo{Err: err, Meta: store})
			return custom_errors.AlreadyExistErr
		}
		//logger.Error(log.LogInfo{Err: err, Info: "create store ошибка в процессе запроса", Meta: store})
		return err
	}

	//logger.Debug(log.LogInfo{Info: "create store завершено", Meta: store})
	return nil
}
