package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StoreRepoPostgres struct {
	db  PgxIface
	log logger.Logger
}

func NewStoreRepoPostgres(db PgxIface, log logger.Logger) *StoreRepoPostgres {
	return &StoreRepoPostgres{
		db:  db,
		log: log,
	}
}

func generateQuery(filter *domain.StoreFilter) (string, []any) {
	query := `
        SELECT DISTINCT s.id, s.name, s.description, s.city_id, s.address, 
               s.card_img, s.rating, s.open_at, s.closed_at,
               array_agg(st.tag_id) as tag_ids
        FROM store s
        LEFT JOIN store_tag st ON s.id = st.store_id
    `
	args := []any{}
	where := []string{}

	// фильтрация по тегу
	if filter.TagID != "" {
		where = append(where, fmt.Sprintf("exists (select 1 from store_tag st2 where st2.store_id = s.id and st2.tag_id = $%d)", len(args)+1))
		args = append(args, filter.TagID)
	}

	// фильтрация по городу
	if filter.CityID != "" {
		where = append(where, fmt.Sprintf("s.city_id = $%d", len(args)+1))
		args = append(args, filter.CityID)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " GROUP BY s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at"

	// если не первый запрос
	if filter.LastID != "" {
		query += fmt.Sprintf(" HAVING s.id > $%d", len(args)+1)
		args = append(args, filter.LastID)
	}

	orderBy := " ORDER BY s.id"
	if filter.Sorted != "" {
		dir := "ASC"
		if filter.Desc {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf(" ORDER BY s.%s %s, s.id", filter.Sorted, dir)
	}
	query += orderBy

	query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, filter.Limit)

	return query, args
}

func (r *StoreRepoPostgres) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.Store, error) {
	r.log.Debug("GetStores начало обработки",
		slog.String("tag_id", filter.TagID),
		slog.String("city_id", filter.CityID),
		slog.String("sorted", filter.Sorted),
		slog.Int("limit", filter.Limit),
	)

	query, args := generateQuery(filter)

	r.log.Debug("Сгенерированный SQL запрос",
		slog.String("query", query),
		slog.Any("args", args),
		slog.Int("args_count", len(args)),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("GetStores ошибка выполнения SQL",
			slog.Any("err", err),
			slog.String("query", query),
			slog.Any("args", args),
		)
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		var tagID *string

		err = rows.Scan(&store.ID, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg, &store.Rating,
			&store.OpenAt, &store.ClosedAt, &tagID)
		if err != nil {
			r.log.Error("GetStores ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}

		if tagID != nil {
			store.TagID = *tagID
		}
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetStores ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(stores) == 0 {
		r.log.Debug("GetStores пустой ответ",
			slog.String("tag_id", filter.TagID),
			slog.String("city_id", filter.CityID),
		)
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetStores завершено успешно", slog.Int("stores_count", len(stores)))
	return stores, nil
}

//go:embed sql/store/get.sql
var getStoreQuery string

func (r *StoreRepoPostgres) GetStore(ctx context.Context, id string) ([]*domain.Store, error) {
	r.log.Debug("GetStore начало обработки", slog.String("id", id))

	rows, err := r.db.Query(ctx, getStoreQuery, id)
	if err != nil {
		r.log.Error("GetStore ошибка бд", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		err = rows.Scan(&store.ID, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg, &store.Rating,
			&store.OpenAt, &store.ClosedAt, &store.TagID)
		if err != nil {
			r.log.Error("GetStore ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetStore ошибка после чтения строк", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}

	if len(stores) == 0 {
		r.log.Debug("GetStore пустой ответ", slog.String("id", id))
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetStore завершено успешно", slog.String("id", id))
	return stores, nil
}

//go:embed sql/store/get_review.sql
var getStoreReview string

func (r *StoreRepoPostgres) GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) {
	r.log.Debug("GetStoreReview начало обработки", slog.String("id", id))

	rows, err := r.db.Query(ctx, getStoreReview, id)
	if err != nil {
		r.log.Error("GetStoreReview ошибка бд", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.StoreReview
	for rows.Next() {
		var review domain.StoreReview
		var createdAt time.Time

		err = rows.Scan(&review.UserName, &review.Rating, &review.Comment, &createdAt)
		if err != nil {
			r.log.Error("GetStoreReview ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}

		review.CreatedAt = createdAt.Format(time.RFC3339)
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetStoreReview ошибка после чтения строк", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}

	if len(reviews) == 0 {
		r.log.Debug("GetStoreReview пустой ответ", slog.String("id", id))
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetStoreReview завершено успешно", slog.String("id", id))
	return reviews, nil
}

//go:embed sql/store/create.sql
var createStore string

func (r *StoreRepoPostgres) CreateStore(ctx context.Context, store *domain.Store) error {
	r.log.Debug("CreateStore начало обработки")

	store.ID = uuid.New().String()
	_, err := r.db.Exec(ctx, createStore, store.ID, store.Name, store.Description,
		store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			r.log.Warn("CreateStore unique ограничение", slog.Any("err", err))
			return domain.ErrStoreExist
		}
		r.log.Error("CreateStore ошибка бд", slog.Any("err", err))
		return err
	}

	r.log.Debug("CreateStore завершено успешно")
	return nil
}

//go:embed sql/store/get_tag.sql
var getTags string

func (r *StoreRepoPostgres) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	r.log.Debug("GetTags начало обработки")

	rows, err := r.db.Query(ctx, getTags)
	if err != nil {
		r.log.Error("GetTags ошибка бд", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.StoreTag
	for rows.Next() {
		var tag domain.StoreTag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			r.log.Error("GetTags ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetTags ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(tags) == 0 {
		r.log.Debug("GetTags пустой ответ")
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetTags завершено успешно")
	return tags, nil
}

//go:embed sql/store/get_city.sql
var getCity string

func (r *StoreRepoPostgres) GetCities(ctx context.Context) ([]*domain.City, error) {
	r.log.Debug("GetCities начало обработки")

	rows, err := r.db.Query(ctx, getCity)
	if err != nil {
		r.log.Error("GetCities ошибка бд", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var cities []*domain.City
	for rows.Next() {
		var city domain.City
		err = rows.Scan(&city.ID, &city.Name)
		if err != nil {
			r.log.Error("GetCities ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetCities ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(cities) == 0 {
		r.log.Debug("GetCities пустой ответ")
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetCities завершено успешно")
	return cities, nil
}
