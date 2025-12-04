package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type StoreRepoPostgres struct {
	db PgxIface
}

func NewStoreRepoPostgres(db PgxIface) *StoreRepoPostgres {
	return &StoreRepoPostgres{
		db: db,
	}
}

func generateQuery(filter *domain.StoreFilter) (string, []any) {
	query := `
    SELECT 
        s.id, s.name, s.description, s.city_id, s.address, 
        s.card_img, s.rating, s.open_at, s.closed_at,
        COALESCE(json_agg(DISTINCT st.tag_id) FILTER (WHERE st.tag_id IS NOT NULL), '[]'::json) AS tag_ids,
        COALESCE(json_agg(DISTINCT sc.category_id) FILTER (WHERE sc.category_id IS NOT NULL), '[]'::json) AS category_ids
    FROM store s
    LEFT JOIN store_tag st ON s.id = st.store_id
    LEFT JOIN store_category sc ON s.id = sc.store_id
`

	args := []any{}
	where := []string{}

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("to_tsvector('russian', s.name || ' ' || s.description) @@ to_tsquery('russian', $%d)", len(args)+1))
		args = append(args, filter.Search)
	}

	// TagIDs как ANY (ИЛИ) - если передано, то найди магазины с ЛЮБЫМ из этих тегов
	if len(filter.TagIDs) > 0 {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM store_tag st2 WHERE st2.store_id = s.id AND st2.tag_id = ANY($%d))", len(args)+1))
		args = append(args, filter.TagIDs)
	}

	// CategoryIDs как AND (И) - должны быть ВСЕ категории
	if len(filter.CategoryIDs) > 0 {
		for _, catID := range filter.CategoryIDs {
			where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM store_category sc2 WHERE sc2.store_id = s.id AND sc2.category_id = $%d)", len(args)+1))
			args = append(args, catID)
		}
	}

	if filter.CityID != "" {
		where = append(where, fmt.Sprintf("s.city_id = $%d", len(args)+1))
		args = append(args, filter.CityID)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " GROUP BY s.id"

	if filter.LastID != "" {
		query += fmt.Sprintf(" HAVING s.id > $%d", len(args)+1)
		args = append(args, filter.LastID)
	}

	orderBy := " ORDER BY s.id"
	if filter.Search != "" {
		orderBy = fmt.Sprintf(" ORDER BY ts_rank(to_tsvector('russian', s.name || ' ' || s.description), to_tsquery('russian', $%d)) DESC, s.id", len(args)+1)
		args = append(args, filter.Search)
	} else if filter.Sorted != "" {
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

func generateSearchWithItemsQuery(filter *domain.StoreSearchFilter) (string, []any) {
	query := `
    SELECT 
        s.id, s.name, s.description, s.city_id, s.address, 
        s.card_img, s.rating, s.open_at, s.closed_at,
        COALESCE(json_agg(DISTINCT st.tag_id) FILTER (WHERE st.tag_id IS NOT NULL), '[]'::json) AS tag_ids,
        COALESCE(json_agg(DISTINCT sc.category_id) FILTER (WHERE sc.category_id IS NOT NULL), '[]'::json) AS category_ids,
        si.id AS item_id,
        i.name AS item_name,
        si.price,
        COALESCE(json_agg(DISTINCT it.type_id) FILTER (WHERE it.type_id IS NOT NULL), '[]'::json) AS item_types
    FROM store s
    LEFT JOIN store_tag st ON s.id = st.store_id
    LEFT JOIN store_category sc ON s.id = sc.store_id
    LEFT JOIN store_item si ON s.id = si.store_id
    LEFT JOIN item i ON si.item_id = i.id
    LEFT JOIN item_type it ON i.id = it.item_id
`

	args := []any{}
	where := []string{}

	if filter.Search != "" {
		paramNum := len(args) + 1
		where = append(where, fmt.Sprintf(`
            (to_tsvector('russian', s.name || ' ' || s.description) @@ to_tsquery('russian', $%d)
            OR to_tsvector('russian', i.name) @@ to_tsquery('russian', $%d))
        `, paramNum, paramNum))
		args = append(args, filter.Search)
	}

	// TagIDs как ANY (ИЛИ)
	if len(filter.TagIDs) > 0 {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM store_tag st2 WHERE st2.store_id = s.id AND st2.tag_id = ANY($%d))", len(args)+1))
		args = append(args, filter.TagIDs)
	}

	// CategoryIDs как AND (И)
	if len(filter.CategoryIDs) > 0 {
		for _, catID := range filter.CategoryIDs {
			where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM store_category sc2 WHERE sc2.store_id = s.id AND sc2.category_id = $%d)", len(args)+1))
			args = append(args, catID)
		}
	}

	if filter.CityID != "" {
		where = append(where, fmt.Sprintf("s.city_id = $%d", len(args)+1))
		args = append(args, filter.CityID)
	}

	if len(filter.ItemTypes) > 0 {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM item_type it2 WHERE it2.item_id = i.id AND it2.type_id = ANY($%d))", len(args)+1))
		args = append(args, filter.ItemTypes)
	}

	if filter.MinPrice > 0 {
		where = append(where, fmt.Sprintf("si.price >= $%d", len(args)+1))
		args = append(args, filter.MinPrice)
	}
	if filter.MaxPrice < 999999 {
		where = append(where, fmt.Sprintf("si.price <= $%d", len(args)+1))
		args = append(args, filter.MaxPrice)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += ` GROUP BY s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at,
                 si.id, i.name, si.price`

	query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, filter.Limit)

	return query, args
}

func (r *StoreRepoPostgres) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetStores начало обработки",
		slog.Any("tag_ids", filter.TagIDs),
		slog.Any("category_ids", filter.CategoryIDs),
		slog.String("city_id", filter.CityID),
		slog.String("sorted", filter.Sorted),
		slog.Int("limit", filter.Limit),
	)

	query, args := generateQuery(filter)

	log.DebugContext(ctx, "Сгенерированный SQL запрос",
		slog.String("query", query),
		slog.Any("args", args),
		slog.Int("args_count", len(args)),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.ErrorContext(ctx, "GetStores ошибка выполнения SQL",
			slog.Any("err", err),
			slog.String("query", query),
			slog.Any("args", args),
		)
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.StoreAgg
	for rows.Next() {
		var store domain.StoreAgg
		var tagIDsJSON string
		var categoryIDsJSON string

		err = rows.Scan(
			&store.ID,
			&store.Name,
			&store.Description,
			&store.CityID,
			&store.Address,
			&store.CardImg,
			&store.Rating,
			&store.OpenAt,
			&store.ClosedAt,
			&tagIDsJSON,
			&categoryIDsJSON,
		)
		if err != nil {
			log.ErrorContext(ctx, "GetStores ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}

		var tagIDs []string
		var categoryIDs []string

		json.Unmarshal([]byte(tagIDsJSON), &tagIDs)
		json.Unmarshal([]byte(categoryIDsJSON), &categoryIDs)

		store.TagsID = tagIDs
		store.CategoriesID = categoryIDs
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetStores ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(stores) == 0 {
		log.DebugContext(ctx, "GetStores пустой результат",
			slog.Any("tag_ids", filter.TagIDs),
			slog.Any("category_ids", filter.CategoryIDs),
			slog.String("city_id", filter.CityID),
		)
		return []*domain.StoreAgg{}, nil
	}

	log.DebugContext(ctx, "GetStores завершено успешно", slog.Int("stores_count", len(stores)))
	return stores, nil
}

func (r *StoreRepoPostgres) SearchStoresWithItems(ctx context.Context, filter *domain.StoreSearchFilter) ([]*domain.StoreWithItems, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "SearchStoresWithItems начало обработки",
		slog.String("search", filter.Search),
		slog.Any("tag_ids", filter.TagIDs),
		slog.Any("category_ids", filter.CategoryIDs),
		slog.Any("item_types", filter.ItemTypes),
		slog.Float64("min_price", filter.MinPrice),
		slog.Float64("max_price", filter.MaxPrice),
	)

	query, args := generateSearchWithItemsQuery(filter)

	log.DebugContext(ctx, "Сгенерированный SQL запрос",
		slog.String("query", query),
		slog.Any("args", args),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.ErrorContext(ctx, "SearchStoresWithItems ошибка выполнения SQL",
			slog.Any("err", err),
			slog.String("query", query),
		)
		return nil, err
	}
	defer rows.Close()

	storesMap := make(map[string]*domain.StoreWithItems)
	var storeOrder []*domain.StoreWithItems

	for rows.Next() {
		var storeID, storeName, description, cityID, address, cardImg, openAt, closedAt string
		var rating float64
		var tagIDsJSON string
		var categoryIDsJSON string
		var itemID sql.NullString
		var itemName sql.NullString
		var price sql.NullFloat64
		var itemTypesJSON string

		err = rows.Scan(
			&storeID,
			&storeName,
			&description,
			&cityID,
			&address,
			&cardImg,
			&rating,
			&openAt,
			&closedAt,
			&tagIDsJSON,
			&categoryIDsJSON,
			&itemID,
			&itemName,
			&price,
			&itemTypesJSON,
		)
		if err != nil {
			log.ErrorContext(ctx, "SearchStoresWithItems ошибка при декодировании", slog.Any("err", err))
			return nil, err
		}

		var tagIDs []string
		var categoryIDs []string
		var itemTypes []string

		json.Unmarshal([]byte(tagIDsJSON), &tagIDs)
		json.Unmarshal([]byte(categoryIDsJSON), &categoryIDs)
		json.Unmarshal([]byte(itemTypesJSON), &itemTypes)

		if _, exists := storesMap[storeID]; !exists {
			store := &domain.StoreWithItems{
				ID:           storeID,
				Name:         storeName,
				Description:  description,
				CityID:       cityID,
				Address:      address,
				CardImg:      cardImg,
				Rating:       rating,
				OpenAt:       openAt,
				ClosedAt:     closedAt,
				TagsID:       tagIDs,
				CategoriesID: categoryIDs,
				Items:        make([]*domain.Item, 0),
			}
			storesMap[storeID] = store
			storeOrder = append(storeOrder, store)
		}

		if itemID.Valid && itemName.Valid && price.Valid {
			item := &domain.Item{
				ID:      itemID.String,
				Name:    itemName.String,
				Price:   price.Float64,
				TypesID: itemTypes,
			}
			storesMap[storeID].Items = append(storesMap[storeID].Items, item)
		}
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "SearchStoresWithItems ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(storeOrder) == 0 {
		log.DebugContext(ctx, "SearchStoresWithItems пустой результат")
		return []*domain.StoreWithItems{}, nil
	}

	log.DebugContext(ctx, "SearchStoresWithItems завершено успешно", slog.Int("stores_count", len(storeOrder)))
	return storeOrder, nil
}

//go:embed sql/store/get.sql
var getStoreQuery string

func (r *StoreRepoPostgres) GetStore(ctx context.Context, id string) (*domain.StoreAgg, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetStore начало обработки", slog.String("id", id))

	rows, err := r.db.Query(ctx, getStoreQuery, id)
	if err != nil {
		log.ErrorContext(ctx, "GetStore ошибка бд", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrRowsNotFound
	}

	var store domain.StoreAgg
	var tagIDsJSON string
	var categoryIDsJSON string

	err = rows.Scan(
		&store.ID,
		&store.Name,
		&store.Description,
		&store.CityID,
		&store.Address,
		&store.CardImg,
		&store.Rating,
		&store.OpenAt,
		&store.ClosedAt,
		&tagIDsJSON,
		&categoryIDsJSON,
	)
	if err != nil {
		log.ErrorContext(ctx, "GetStore ошибка при декодировании данных", slog.Any("err", err))
		return nil, err
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetStore ошибка после чтения строк", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}

	var tagIDs []string
	var categoryIDs []string

	json.Unmarshal([]byte(tagIDsJSON), &tagIDs)
	json.Unmarshal([]byte(categoryIDsJSON), &categoryIDs)

	store.TagsID = tagIDs
	store.CategoriesID = categoryIDs
	log.DebugContext(ctx, "GetStore завершено успешно", slog.String("id", id))
	return &store, nil
}

//go:embed sql/store/get_review.sql
var getStoreReview string

func (r *StoreRepoPostgres) GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetStoreReview начало обработки", slog.String("id", id))

	rows, err := r.db.Query(ctx, getStoreReview, id)
	if err != nil {
		log.ErrorContext(ctx, "GetStoreReview ошибка бд", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.StoreReview
	for rows.Next() {
		var review domain.StoreReview
		var createdAt time.Time

		err = rows.Scan(&review.UserName, &review.Rating, &review.Comment, &createdAt)
		if err != nil {
			log.ErrorContext(ctx, "GetStoreReview ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}

		review.CreatedAt = createdAt.Format(time.RFC3339)
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetStoreReview ошибка после чтения строк", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}

	if len(reviews) == 0 {
		log.DebugContext(ctx, "GetStoreReview пустой ответ", slog.String("id", id))
		return []*domain.StoreReview{}, nil
	}

	log.DebugContext(ctx, "GetStoreReview завершено успешно", slog.String("id", id))
	return reviews, nil
}

//go:embed sql/store/create.sql
var createStore string

func (r *StoreRepoPostgres) CreateStore(ctx context.Context, store *domain.Store) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "CreateStore начало обработки")

	store.ID = uuid.New().String()
	_, err := r.db.Exec(ctx, createStore, store.ID, store.Name, store.Description,
		store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.WarnContext(ctx, "CreateStore unique ограничение", slog.Any("err", err))
			return domain.ErrStoreExist
		}
		log.ErrorContext(ctx, "CreateStore ошибка бд", slog.Any("err", err))
		return err
	}

	log.DebugContext(ctx, "CreateStore завершено успешно")
	return nil
}

//go:embed sql/store/get_tag.sql
var getTags string

func (r *StoreRepoPostgres) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetTags начало обработки")

	rows, err := r.db.Query(ctx, getTags)
	if err != nil {
		log.ErrorContext(ctx, "GetTags ошибка бд", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.StoreTag
	for rows.Next() {
		var tag domain.StoreTag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetTags ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetTags ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(tags) == 0 {
		log.DebugContext(ctx, "GetTags пустой ответ")
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetTags завершено успешно", slog.Int("tags_count", len(tags)))
	return tags, nil
}

//go:embed sql/store/get_category.sql
var getCategories string

func (r *StoreRepoPostgres) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetCategories начало обработки")

	rows, err := r.db.Query(ctx, getCategories)
	if err != nil {
		log.ErrorContext(ctx, "GetCategories ошибка бд", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var category domain.Category
		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetCategories ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetCategories ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(categories) == 0 {
		log.DebugContext(ctx, "GetCategories пустой ответ")
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetCategories завершено успешно", slog.Int("categories_count", len(categories)))
	return categories, nil
}

//go:embed sql/store/get_city.sql
var getCity string

func (r *StoreRepoPostgres) GetCities(ctx context.Context) ([]*domain.City, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetCities начало обработки")

	rows, err := r.db.Query(ctx, getCity)
	if err != nil {
		log.ErrorContext(ctx, "GetCities ошибка бд", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var cities []*domain.City
	for rows.Next() {
		var city domain.City
		err = rows.Scan(&city.ID, &city.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetCities ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetCities ошибка после чтения строк", slog.Any("err", err))
		return nil, err
	}

	if len(cities) == 0 {
		log.DebugContext(ctx, "GetCities пустой ответ")
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetCities завершено успешно", slog.Int("cities_count", len(cities)))
	return cities, nil
}
