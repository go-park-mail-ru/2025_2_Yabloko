package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	    select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at
    from store s
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

	// если не первый запрос
	if filter.LastID != "" {
		where = append(where, fmt.Sprintf("s.id > $%d", len(args)+1))
		args = append(args, filter.LastID)
	}

	if len(where) > 0 {
		query += " where " + strings.Join(where, " and ")
	}

	orderBy := " order by s.id"
	if filter.Sorted != "" {
		dir := "asc"
		if filter.Desc {
			dir = "desc"
		}
		orderBy = fmt.Sprintf(" order by s.%s %s, s.id", filter.Sorted, dir)
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

//go:embed sql/store/get.sql
var getStoreQuery string

func (r *StoreRepoPostgres) GetStore(ctx context.Context, id string) ([]*domain.Store, error) {
	r.log.Debug(ctx, "GetStore начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getStoreQuery, id)
	if err != nil {
		r.log.Error(ctx, "GetStore ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		err = rows.Scan(&store.ID, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt, &store.TagID)
		if err != nil {
			r.log.Error(ctx, "GetStore ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetStore ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	if len(stores) == 0 {
		r.log.Debug(ctx, "GetStore пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetStore завершено успешно", map[string]interface{}{})
	return stores, nil
}

//go:embed sql/store/get_review.sql
var getStoreReview string

func (r *StoreRepoPostgres) GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) {
	r.log.Debug(ctx, "GetStoreReview начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getStoreReview, id)
	if err != nil {
		r.log.Error(ctx, "GetStoreReview ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.StoreReview
	for rows.Next() {
		var review domain.StoreReview
		err = rows.Scan(&review.UserName, &review.Rating, &review.Comment, &review.CreatedAt)
		if err != nil {
			r.log.Error(ctx, "GetStoreReview ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetStoreReview ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	if len(reviews) == 0 {
		r.log.Debug(ctx, "GetStoreReview пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetStoreReview завершено успешно", map[string]interface{}{})
	return reviews, nil
}

//go:embed sql/store/create.sql
var createStore string

// CreateStore не используется
func (r *StoreRepoPostgres) CreateStore(ctx context.Context, store *domain.Store) error {
	r.log.Debug(ctx, "CreateStore начало обработки", map[string]interface{}{})

	store.ID = uuid.New().String()
	_, err := r.db.Exec(ctx, createStore, store.ID, store.Name, store.Description,
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

//go:embed sql/store/get_tag.sql
var getTags string

func (r *StoreRepoPostgres) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	r.log.Debug(ctx, "GetTags начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getTags)
	if err != nil {
		r.log.Error(ctx, "GetTags ошибка бд", map[string]interface{}{"err": err})
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.StoreTag
	for rows.Next() {
		var tag domain.StoreTag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			r.log.Error(ctx, "GetTags ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetTags ошибка после чтения строк",
			map[string]interface{}{"err": err})
		return nil, err
	}

	if len(tags) == 0 {
		r.log.Debug(ctx, "GetTags пустой ответ", map[string]interface{}{})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetTags завершено успешно", map[string]interface{}{})
	return tags, nil
}

//go:embed sql/store/get_city.sql
var getCity string

func (r *StoreRepoPostgres) GetCities(ctx context.Context) ([]*domain.City, error) {
	r.log.Debug(ctx, "GetCities начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getCity)
	if err != nil {
		r.log.Error(ctx, "GetCities ошибка бд", map[string]interface{}{"err": err})
		return nil, err
	}
	defer rows.Close()

	var cities []*domain.City
	for rows.Next() {
		var city domain.City
		err = rows.Scan(&city.ID, &city.Name)
		if err != nil {
			r.log.Error(ctx, "GetCities ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetCities ошибка после чтения строк",
			map[string]interface{}{"err": err})
		return nil, err
	}

	if len(cities) == 0 {
		r.log.Debug(ctx, "GetCities пустой ответ", map[string]interface{}{})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetCities завершено успешно", map[string]interface{}{})
	return cities, nil
}
