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
	"github.com/pgvector/pgvector-go"
)

//go:embed sql/store/get.sql
var getStoreQuery string

//go:embed sql/store/get_stores.sql
var getStoresQuery string

//go:embed sql/store/search_with_items.sql
var searchWithItemsQuery string

//go:embed sql/store/search_hybrid.sql
var searchHybridQuery string

//go:embed sql/store/create.sql
var createStoreQuery string

//go:embed sql/store/get_review.sql
var getStoreReviewQuery string

//go:embed sql/store/update_embedding.sql
var updateStoreEmbeddingQuery string

//go:embed sql/store/get_tag.sql
var getTagsQuery string

//go:embed sql/store/get_category.sql
var getCategoriesQuery string

//go:embed sql/store/get_city.sql
var getCitiesQuery string

//go:embed sql/store/get_stores_without_embedding.sql
var getStoresWithoutEmbeddingQuery string

type StoreRepoPostgres struct {
	db PgxIface
}

func NewStoreRepoPostgres(db PgxIface) *StoreRepoPostgres {
	return &StoreRepoPostgres{db: db}
}

type QueryBuilder struct {
	whereConditions []string
	args            []any
	paramCount      int
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		whereConditions: []string{},
		args:            []any{},
		paramCount:      0,
	}
}

func (qb *QueryBuilder) AddCondition(condition string, args ...any) {
	for _, arg := range args {
		qb.paramCount++
		qb.args = append(qb.args, arg)
	}
	placeholderCondition := condition
	for i := 1; i <= len(args); i++ {
		placeholderCondition = strings.Replace(
			placeholderCondition,
			fmt.Sprintf("$%d", i),
			fmt.Sprintf("$%d", qb.paramCount-len(args)+i),
			1,
		)
	}
	qb.whereConditions = append(qb.whereConditions, placeholderCondition)
}

func (qb *QueryBuilder) BuildWhere() (string, []any) {
	if len(qb.whereConditions) == 0 {
		return "", qb.args
	}
	return " WHERE " + strings.Join(qb.whereConditions, " AND "), qb.args
}

func (qb *QueryBuilder) GetNextParam() int {
	return qb.paramCount + 1
}

func (r *StoreRepoPostgres) GetStores(ctx context.Context, filter *domain.StoreFilter) ([]*domain.StoreAgg, error) {
	log := logger.FromContext(ctx)

	qb := NewQueryBuilder()

	if filter.Search != "" {
		qb.AddCondition(`to_tsvector('russian', s.name || ' ' || s.description) @@ to_tsquery('russian', $1)`, filter.Search)
	}

	if len(filter.TagIDs) > 0 {
		for _, tagID := range filter.TagIDs {
			qb.AddCondition(`EXISTS (SELECT 1 FROM store_tag st2 WHERE st2.store_id = s.id AND st2.tag_id = $1)`, tagID)
		}
	}

	if len(filter.CategoryIDs) > 0 {
		placeholders := make([]string, len(filter.CategoryIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM store_category sc2 WHERE sc2.store_id = s.id AND sc2.category_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, catID := range filter.CategoryIDs {
			qb.args = append(qb.args, catID)
		}
		qb.paramCount += len(filter.CategoryIDs)
	}

	if filter.CityID != "" {
		qb.AddCondition(`s.city_id = $1`, filter.CityID)
	}

	where, args := qb.BuildWhere()
	query := getStoresQuery + where + " GROUP BY s.id"

	orderBy := " ORDER BY s.id"
	if filter.Search != "" {
		orderBy = fmt.Sprintf(` ORDER BY ts_rank(to_tsvector('russian', s.name || ' ' || s.description), to_tsquery('russian', $%d)) DESC, s.id`, len(args)+1)
		args = append(args, filter.Search)
	} else if filter.Sorted != "" {
		allowedSorts := map[string]bool{"rating": true, "open_at": true, "closed_at": true}
		if allowedSorts[filter.Sorted] {
			dir := "ASC"
			if filter.Desc {
				dir = "DESC"
			}
			orderBy = fmt.Sprintf(" ORDER BY s.%s %s, s.id", filter.Sorted, dir)
		}
	}
	query += orderBy

	if filter.LastID != "" {
		query += fmt.Sprintf(" HAVING s.id > $%d", len(args)+1)
		args = append(args, filter.LastID)
	}

	query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, filter.Limit)

	log.DebugContext(ctx, "GetStores",
		slog.Int("tag_ids_count", len(filter.TagIDs)),
		slog.Int("category_ids_count", len(filter.CategoryIDs)),
		slog.String("city_id", filter.CityID),
		slog.String("search", filter.Search),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.ErrorContext(ctx, "GetStores query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.StoreAgg
	for rows.Next() {
		var store domain.StoreAgg
		var tagIDsJSON, categoryIDsJSON string

		err = rows.Scan(
			&store.ID, &store.Name, &store.Description, &store.CityID,
			&store.Address, &store.CardImg, &store.Rating, &store.OpenAt,
			&store.ClosedAt, &tagIDsJSON, &categoryIDsJSON,
		)
		if err != nil {
			log.ErrorContext(ctx, "GetStores scan error", slog.Any("err", err))
			return nil, err
		}

		var tagIDs, categoryIDs []string
		json.Unmarshal([]byte(tagIDsJSON), &tagIDs)
		json.Unmarshal([]byte(categoryIDsJSON), &categoryIDs)

		store.TagsID = tagIDs
		store.CategoriesID = categoryIDs
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetStores rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "GetStores completed", slog.Int("count", len(stores)))
	return stores, nil
}

func (r *StoreRepoPostgres) SearchStoresWithItems(ctx context.Context, filter *domain.StoreSearchFilter) ([]*domain.StoreWithItems, error) {
	log := logger.FromContext(ctx)

	qb := NewQueryBuilder()

	if filter.Search != "" {
		qb.AddCondition(
			`(to_tsvector('russian', s.name || ' ' || s.description) @@ to_tsquery('russian', $1) OR to_tsvector('russian', i.name) @@ to_tsquery('russian', $2))`,
			filter.Search, filter.Search,
		)
	}

	if len(filter.TagIDs) > 0 {
		placeholders := make([]string, len(filter.TagIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM store_tag st2 WHERE st2.store_id = s.id AND st2.tag_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, tagID := range filter.TagIDs {
			qb.args = append(qb.args, tagID)
		}
		qb.paramCount += len(filter.TagIDs)
	}

	if len(filter.CategoryIDs) > 0 {
		placeholders := make([]string, len(filter.CategoryIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM store_category sc2 WHERE sc2.store_id = s.id AND sc2.category_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, catID := range filter.CategoryIDs {
			qb.args = append(qb.args, catID)
		}
		qb.paramCount += len(filter.CategoryIDs)
	}

	if filter.CityID != "" {
		qb.AddCondition(fmt.Sprintf(`s.city_id = $%d`, qb.GetNextParam()), filter.CityID)
	}

	if len(filter.ItemTypes) > 0 {
		placeholders := make([]string, len(filter.ItemTypes))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM item_type it2 WHERE it2.item_id = i.id AND it2.type_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, typeID := range filter.ItemTypes {
			qb.args = append(qb.args, typeID)
		}
		qb.paramCount += len(filter.ItemTypes)
	}

	if filter.MinPrice > 0 {
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`si.price >= $%d`, qb.paramCount+1))
		qb.args = append(qb.args, filter.MinPrice)
		qb.paramCount++
	}

	if filter.MaxPrice < 999999 {
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`si.price <= $%d`, qb.paramCount+1))
		qb.args = append(qb.args, filter.MaxPrice)
		qb.paramCount++
	}

	where := ""
	if len(qb.whereConditions) > 0 {
		where = " WHERE " + strings.Join(qb.whereConditions, " AND ")
	}

	query := searchWithItemsQuery + where + `
	GROUP BY s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at,
	         si.id, i.name, si.price, i.embedding
	`

	query += fmt.Sprintf(" LIMIT $%d", len(qb.args)+1)
	queryArgs := append(qb.args, filter.Limit)

	log.DebugContext(ctx, "SearchStoresWithItems", slog.String("search", filter.Search))

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		log.ErrorContext(ctx, "SearchStoresWithItems query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	storesMap := make(map[string]*domain.StoreWithItems)
	var storeOrder []*domain.StoreWithItems

	for rows.Next() {
		var storeID, storeName, description, cityID, address, cardImg, openAt, closedAt string
		var rating float64
		var tagIDsJSON, categoryIDsJSON string
		var itemID sql.NullString
		var itemName sql.NullString
		var price sql.NullFloat64
		var itemEmbedding sql.NullString
		var itemTypesJSON string

		err = rows.Scan(
			&storeID, &storeName, &description, &cityID, &address, &cardImg,
			&rating, &openAt, &closedAt,
			&tagIDsJSON, &categoryIDsJSON,
			&itemID, &itemName, &price, &itemEmbedding, &itemTypesJSON,
		)
		if err != nil {
			log.ErrorContext(ctx, "SearchStoresWithItems scan error", slog.Any("err", err))
			return nil, err
		}

		var tagIDs, categoryIDs, itemTypes []string
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
		log.ErrorContext(ctx, "SearchStoresWithItems rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "SearchStoresWithItems completed", slog.Int("count", len(storeOrder)))
	return storeOrder, nil
}

func (r *StoreRepoPostgres) SearchStoresHybrid(
	ctx context.Context,
	filter *domain.StoreSearchFilter,
	embedding []float32,
) ([]*domain.StoreWithItems, error) {
	log := logger.FromContext(ctx)

	qb := NewQueryBuilder()

	args := []any{
		filter.Search,
		pgvector.NewVector(embedding),
		0.6,
		0.4,
	}
	qb.paramCount = 4
	qb.args = args

	if len(filter.TagIDs) > 0 {
		placeholders := make([]string, len(filter.TagIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM store_tag st2 WHERE st2.store_id = s.id AND st2.tag_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, tagID := range filter.TagIDs {
			qb.args = append(qb.args, tagID)
		}
		qb.paramCount += len(filter.TagIDs)
	}

	if len(filter.CategoryIDs) > 0 {
		placeholders := make([]string, len(filter.CategoryIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM store_category sc2 WHERE sc2.store_id = s.id AND sc2.category_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, catID := range filter.CategoryIDs {
			qb.args = append(qb.args, catID)
		}
		qb.paramCount += len(filter.CategoryIDs)
	}

	if filter.CityID != "" {
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf("s.city_id = $%d", qb.paramCount+1))
		qb.args = append(qb.args, filter.CityID)
		qb.paramCount++
	}

	if len(filter.ItemTypes) > 0 {
		placeholders := make([]string, len(filter.ItemTypes))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM item_type it2 WHERE it2.item_id = i.id AND it2.type_id = ANY(ARRAY[%s]::uuid[]))`, strings.Join(placeholders, ",")))
		for _, typeID := range filter.ItemTypes {
			qb.args = append(qb.args, typeID)
		}
		qb.paramCount += len(filter.ItemTypes)
	}

	if filter.MinPrice > 0 {
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf("si.price >= $%d", qb.paramCount+1))
		qb.args = append(qb.args, filter.MinPrice)
		qb.paramCount++
	}

	if filter.MaxPrice < 999999 {
		qb.whereConditions = append(qb.whereConditions, fmt.Sprintf("si.price <= $%d", qb.paramCount+1))
		qb.args = append(qb.args, filter.MaxPrice)
		qb.paramCount++
	}

	where := ""
	if len(qb.whereConditions) > 0 {
		where = " WHERE " + strings.Join(qb.whereConditions, " AND ")
	}

	query := searchHybridQuery + where + fmt.Sprintf(` LIMIT $%d`, len(qb.args)+1)

	queryArgs := append(qb.args, filter.Limit)

	log.DebugContext(ctx, "SearchStoresHybrid",
		slog.String("search", filter.Search),
		slog.Int("embedding_dim", len(embedding)),
	)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		log.ErrorContext(ctx, "SearchStoresHybrid query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	storesMap := make(map[string]*domain.StoreWithItems)
	var storeOrder []*domain.StoreWithItems

	for rows.Next() {
		var storeID, storeName, description, cityID, address, cardImg, openAt, closedAt string
		var rating float64
		var tagIDsJSON, categoryIDsJSON string
		var itemID sql.NullString
		var itemName sql.NullString
		var price sql.NullFloat64
		var itemEmbedding sql.NullString
		var itemTypesJSON string

		err = rows.Scan(
			&storeID, &storeName, &description, &cityID, &address, &cardImg,
			&rating, &openAt, &closedAt,
			&tagIDsJSON, &categoryIDsJSON,
			&itemID, &itemName, &price, &itemEmbedding, &itemTypesJSON,
		)
		if err != nil {
			log.ErrorContext(ctx, "SearchStoresHybrid scan error", slog.Any("err", err))
			return nil, err
		}

		var tagIDs, categoryIDs, itemTypes []string
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
		log.ErrorContext(ctx, "SearchStoresHybrid rows error", slog.Any("err", err))
		return nil, err
	}

	log.InfoContext(ctx, "SearchStoresHybrid completed", slog.Int("count", len(storeOrder)))
	return storeOrder, nil
}

func (r *StoreRepoPostgres) GetStore(ctx context.Context, id string) (*domain.StoreAgg, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetStore", slog.String("id", id))

	rows, err := r.db.Query(ctx, getStoreQuery, id)
	if err != nil {
		log.ErrorContext(ctx, "GetStore query failed", slog.Any("err", err), slog.String("id", id))
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrRowsNotFound
	}

	var store domain.StoreAgg
	var tagIDsJSON, categoryIDsJSON string

	err = rows.Scan(
		&store.ID, &store.Name, &store.Description, &store.CityID,
		&store.Address, &store.CardImg, &store.Rating, &store.OpenAt,
		&store.ClosedAt, &tagIDsJSON, &categoryIDsJSON,
	)
	if err != nil {
		log.ErrorContext(ctx, "GetStore scan error", slog.Any("err", err))
		return nil, err
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetStore rows error", slog.Any("err", err))
		return nil, err
	}

	var tagIDs, categoryIDs []string
	json.Unmarshal([]byte(tagIDsJSON), &tagIDs)
	json.Unmarshal([]byte(categoryIDsJSON), &categoryIDs)

	store.TagsID = tagIDs
	store.CategoriesID = categoryIDs

	log.DebugContext(ctx, "GetStore completed", slog.String("id", id))
	return &store, nil
}

func (r *StoreRepoPostgres) GetStoreReview(ctx context.Context, id string) ([]*domain.StoreReview, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetStoreReview", slog.String("id", id))

	rows, err := r.db.Query(ctx, getStoreReviewQuery, id)
	if err != nil {
		log.ErrorContext(ctx, "GetStoreReview query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.StoreReview
	for rows.Next() {
		var review domain.StoreReview
		var createdAt time.Time

		err = rows.Scan(&review.UserName, &review.Rating, &review.Comment, &createdAt)
		if err != nil {
			log.ErrorContext(ctx, "GetStoreReview scan error", slog.Any("err", err))
			return nil, err
		}

		review.CreatedAt = createdAt.Format(time.RFC3339)
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetStoreReview rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "GetStoreReview completed", slog.Int("count", len(reviews)))
	return reviews, nil
}

func (r *StoreRepoPostgres) CreateStore(ctx context.Context, store *domain.Store) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "CreateStore", slog.String("name", store.Name))

	store.ID = uuid.New().String()
	_, err := r.db.Exec(ctx, createStoreQuery,
		store.ID, store.Name, store.Description,
		store.CityID, store.Address, store.CardImg,
		store.Rating, store.OpenAt, store.ClosedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.WarnContext(ctx, "CreateStore unique constraint violation")
			return domain.ErrStoreExist
		}
		log.ErrorContext(ctx, "CreateStore query failed", slog.Any("err", err))
		return err
	}

	log.DebugContext(ctx, "CreateStore completed", slog.String("id", store.ID))
	return nil
}

func (r *StoreRepoPostgres) UpdateStoreEmbedding(ctx context.Context, storeID string, embedding []float32) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "UpdateStoreEmbedding", slog.String("id", storeID))

	vec := pgvector.NewVector(embedding)
	_, err := r.db.Exec(ctx, updateStoreEmbeddingQuery, storeID, vec)
	if err != nil {
		log.ErrorContext(ctx, "UpdateStoreEmbedding query failed", slog.Any("err", err))
		return err
	}

	log.DebugContext(ctx, "UpdateStoreEmbedding completed", slog.String("id", storeID))
	return nil
}

func (r *StoreRepoPostgres) GetTags(ctx context.Context) ([]*domain.StoreTag, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetTags")

	rows, err := r.db.Query(ctx, getTagsQuery)
	if err != nil {
		log.ErrorContext(ctx, "GetTags query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.StoreTag
	for rows.Next() {
		var tag domain.StoreTag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetTags scan error", slog.Any("err", err))
			return nil, err
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetTags rows error", slog.Any("err", err))
		return nil, err
	}

	if len(tags) == 0 {
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetTags completed", slog.Int("count", len(tags)))
	return tags, nil
}

func (r *StoreRepoPostgres) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetCategories")

	rows, err := r.db.Query(ctx, getCategoriesQuery)
	if err != nil {
		log.ErrorContext(ctx, "GetCategories query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var category domain.Category
		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetCategories scan error", slog.Any("err", err))
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetCategories rows error", slog.Any("err", err))
		return nil, err
	}

	if len(categories) == 0 {
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetCategories completed", slog.Int("count", len(categories)))
	return categories, nil
}

func (r *StoreRepoPostgres) GetCities(ctx context.Context) ([]*domain.City, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetCities")

	rows, err := r.db.Query(ctx, getCitiesQuery)
	if err != nil {
		log.ErrorContext(ctx, "GetCities query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var cities []*domain.City
	for rows.Next() {
		var city domain.City
		err = rows.Scan(&city.ID, &city.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetCities scan error", slog.Any("err", err))
			return nil, err
		}
		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetCities rows error", slog.Any("err", err))
		return nil, err
	}

	if len(cities) == 0 {
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetCities completed", slog.Int("count", len(cities)))
	return cities, nil
}

func (r *StoreRepoPostgres) GetStoresWithoutEmbedding(ctx context.Context) ([]*domain.Store, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetStoresWithoutEmbedding")

	rows, err := r.db.Query(ctx, getStoresWithoutEmbeddingQuery)
	if err != nil {
		log.ErrorContext(ctx, "query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		err = rows.Scan(
			&store.ID, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg,
			&store.Rating, &store.OpenAt, &store.ClosedAt,
		)
		if err != nil {
			log.ErrorContext(ctx, "scan error", slog.Any("err", err))
			return nil, err
		}
		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "GetStoresWithoutEmbedding completed", slog.Int("count", len(stores)))
	return stores, nil
}
