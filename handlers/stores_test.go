package handlers

import (
	"apple_backend/custom_errors"
	"apple_backend/db"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

type storeArgs struct {
	params    db.GetRequest
	setupPool func(pool pgxmock.PgxPoolIface)
	request   *http.Request
}

type storeTestCase struct {
	name           string
	args           storeArgs
	expectedResult []db.ResponseInfo
	responseCode   int
}

// parseJSON ТОЛЬКО ДЛЯ ТЕСТОВ
func parseJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestGetStoresLimit(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := NewHandler(mockPool, "", "", 0)
	router := NewStoresRouter(h)

	storesData := [][]any{
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
			"Суши Мастер",
			"Современный японский ресторан с широким выбором суши и роллов.",
			uuid.MustParse("00000000-0000-0000-0000-000000000011").String(),
			"ул. Тверская, 12",
			"/stores/1.jpg",
			4.5,
			"10:00",
			"23:00",
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			"Sushi House",
			"Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			uuid.MustParse("00000000-0000-0000-0000-000000000012").String(),
			"Невский пр., 45",
			"/stores/2.jpg",
			4.3,
			"10:00",
			"23:00",
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000003").String(),
			"Pizza Roma",
			"Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			uuid.MustParse("00000000-0000-0000-0000-000000000013").String(),
			"ул. Ленина, 78",
			"/stores/3.jpg",
			4.6,
			"10:00",
			"23:00",
		},
	}

	expected := []db.ResponseInfo{
		{
			ID:          "00000000-0000-0000-0000-000000000001",
			Name:        "Суши Мастер",
			Description: "Современный японский ресторан с широким выбором суши и роллов.",
			CityID:      "00000000-0000-0000-0000-000000000011",
			Address:     "ул. Тверская, 12",
			CardImg:     "/stores/1.jpg",
			Rating:      4.5,
			OpenAt:      "10:00",
			ClosedAt:    "23:00",
		},
		{
			ID:          "00000000-0000-0000-0000-000000000002",
			Name:        "Sushi House",
			Description: "Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			CityID:      "00000000-0000-0000-0000-000000000012",
			Address:     "Невский пр., 45",
			CardImg:     "/stores/2.jpg",
			Rating:      4.3,
			OpenAt:      "10:00",
			ClosedAt:    "23:00",
		},
		{
			ID:          "00000000-0000-0000-0000-000000000003",
			Name:        "Pizza Roma",
			Description: "Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			CityID:      "00000000-0000-0000-0000-000000000013",
			Address:     "ул. Ленина, 78",
			CardImg:     "/stores/3.jpg",
			Rating:      4.6,
			OpenAt:      "10:00",
			ClosedAt:    "23:00",
		},
	}

	cases := []storeTestCase{
		{
			name: "Limit=1",
			args: storeArgs{
				params: db.GetRequest{Limit: 1},
				setupPool: func(pool pgxmock.PgxPoolIface) {
					rows := pgxmock.NewRows([]string{
						"id", "name", "description", "city_id", "address", "card_img",
						"rating", "open_at", "closed_at",
					}).AddRow(storesData[0]...)

					pool.ExpectQuery(regexp.QuoteMeta(`
						select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
						from store
						order by id
						limit $1
					`)).
						WithArgs(1).
						WillReturnRows(rows)
				},
				request: httptest.NewRequest(http.MethodPost, "/stores",
					bytes.NewBuffer(parseJSON(db.GetRequest{Limit: 1}))),
			},
			expectedResult: expected[:1],
			responseCode:   http.StatusOK,
		},
		{
			name: "Limit=2",
			args: storeArgs{
				params: db.GetRequest{Limit: 2},
				setupPool: func(pool pgxmock.PgxPoolIface) {
					rows := pgxmock.NewRows([]string{
						"id", "name", "description", "city_id", "address", "card_img",
						"rating", "open_at", "closed_at",
					}).AddRow(storesData[0]...).AddRow(storesData[1]...)

					pool.ExpectQuery(regexp.QuoteMeta(`
						select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
						from store
						order by id
						limit $1
					`)).
						WithArgs(2).
						WillReturnRows(rows)
				},
				request: httptest.NewRequest(http.MethodPost, "/stores",
					bytes.NewBuffer(parseJSON(db.GetRequest{Limit: 2}))),
			},
			expectedResult: expected[:2],
			responseCode:   http.StatusOK,
		},
		{
			name: "Limit=3",
			args: storeArgs{
				params: db.GetRequest{Limit: 3},
				setupPool: func(pool pgxmock.PgxPoolIface) {
					rows := pgxmock.NewRows([]string{
						"id", "name", "description", "city_id", "address", "card_img",
						"rating", "open_at", "closed_at",
					}).AddRow(storesData[0]...).AddRow(storesData[1]...).AddRow(storesData[2]...)

					pool.ExpectQuery(regexp.QuoteMeta(`
						select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
						from store
						order by id
						limit $1
					`)).
						WithArgs(3).
						WillReturnRows(rows)
				},
				request: httptest.NewRequest(http.MethodPost, "/stores",
					bytes.NewBuffer(parseJSON(db.GetRequest{Limit: 3}))),
			},
			expectedResult: expected[:3],
			responseCode:   http.StatusOK,
		},
		{
			name: "Limit=4",
			args: storeArgs{
				params: db.GetRequest{Limit: 4},
				setupPool: func(pool pgxmock.PgxPoolIface) {
					rows := pgxmock.NewRows([]string{
						"id", "name", "description", "city_id", "address", "card_img",
						"rating", "open_at", "closed_at",
					}).AddRow(storesData[0]...).AddRow(storesData[1]...).AddRow(storesData[2]...)

					pool.ExpectQuery(regexp.QuoteMeta(`
						select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
						from store
						order by id
						limit $1
					`)).
						WithArgs(4).
						WillReturnRows(rows)
				},
				request: httptest.NewRequest(http.MethodPost, "/stores",
					bytes.NewBuffer(parseJSON(db.GetRequest{Limit: 4}))),
			},
			expectedResult: expected[:3],
			responseCode:   http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setupPool(mockPool)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.args.request)

			require.Equal(t, tt.responseCode, w.Code)

			expectedJSON, _ := json.Marshal(tt.expectedResult)
			require.JSONEq(t, string(expectedJSON), w.Body.String())
		})
	}
}

func TestGetStoresLimitLastID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := NewHandler(mockPool, "", "", 0)
	router := NewStoresRouter(h)

	storesData := [][]any{
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
			"Суши Мастер",
			"Современный японский ресторан с широким выбором суши и роллов.",
			uuid.MustParse("00000000-0000-0000-0000-000000000011").String(),
			"ул. Тверская, 12",
			"/stores/1.jpg",
			4.5,
			"10:00",
			"23:00",
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			"Sushi House",
			"Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			uuid.MustParse("00000000-0000-0000-0000-000000000012").String(),
			"Невский пр., 45",
			"/stores/2.jpg",
			4.3,
			"10:00",
			"23:00",
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000003").String(),
			"Pizza Roma",
			"Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			uuid.MustParse("00000000-0000-0000-0000-000000000013").String(),
			"ул. Ленина, 78",
			"/stores/3.jpg",
			4.6,
			"10:00",
			"23:00",
		},
	}

	expected := []db.ResponseInfo{
		{
			ID:          "00000000-0000-0000-0000-000000000002",
			Name:        "Sushi House",
			Description: "Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			CityID:      "00000000-0000-0000-0000-000000000012",
			Address:     "Невский пр., 45",
			CardImg:     "/stores/2.jpg",
			Rating:      4.3,
			OpenAt:      "10:00",
			ClosedAt:    "23:00",
		},
		{
			ID:          "00000000-0000-0000-0000-000000000003",
			Name:        "Pizza Roma",
			Description: "Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			CityID:      "00000000-0000-0000-0000-000000000013",
			Address:     "ул. Ленина, 78",
			CardImg:     "/stores/3.jpg",
			Rating:      4.6,
			OpenAt:      "10:00",
			ClosedAt:    "23:00",
		},
	}

	cases := []storeTestCase{
		{
			name: "Limit=1 with ID1",
			args: storeArgs{
				params: db.GetRequest{Limit: 1, LastId: "00000000-0000-0000-0000-000000000001"},
				setupPool: func(pool pgxmock.PgxPoolIface) {
					rows := pgxmock.NewRows([]string{
						"id", "name", "description", "city_id", "address", "card_img",
						"rating", "open_at", "closed_at",
					}).AddRow(storesData[1]...)

					pool.ExpectQuery(regexp.QuoteMeta(`
						select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
						from store
						where id > $1
						order by id
						limit $2
					`)).
						WithArgs("00000000-0000-0000-0000-000000000001", 1).
						WillReturnRows(rows)
				},
				request: httptest.NewRequest(http.MethodPost, "/stores",
					bytes.NewBuffer(parseJSON(db.GetRequest{Limit: 1, LastId: "00000000-0000-0000-0000-000000000001"}))),
			},
			expectedResult: []db.ResponseInfo{expected[0]},
			responseCode:   http.StatusOK,
		},
		{
			name: "Limit=1 with ID3",
			args: storeArgs{
				params: db.GetRequest{Limit: 1, LastId: "00000000-0000-0000-0000-000000000003"},
				setupPool: func(pool pgxmock.PgxPoolIface) {
					rows := pgxmock.NewRows([]string{
						"id", "name", "description", "city_id", "address", "card_img",
						"rating", "open_at", "closed_at",
					})

					pool.ExpectQuery(regexp.QuoteMeta(`
						select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
						from store
						where id > $1
						order by id
						limit $2
					`)).
						WithArgs("00000000-0000-0000-0000-000000000003", 1).
						WillReturnRows(rows)
				},
				request: httptest.NewRequest(http.MethodPost, "/stores",
					bytes.NewBuffer(parseJSON(db.GetRequest{Limit: 1, LastId: "00000000-0000-0000-0000-000000000003"}))),
			},
			expectedResult: []db.ResponseInfo{},
			responseCode:   http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setupPool(mockPool)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.args.request)

			require.Equal(t, tt.responseCode, w.Code)

			expectedJSON, _ := json.Marshal(tt.expectedResult)
			require.JSONEq(t, string(expectedJSON), w.Body.String())
		})
	}
}

func TestGetStoresNegativeLimit(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := NewHandler(mockPool, "", "", 0)
	mux := NewStoresRouter(h)

	body, _ := json.Marshal(db.GetRequest{Limit: -1})
	req := httptest.NewRequest(http.MethodPost, "/stores", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	expectedJSON := fmt.Sprintf(`{"error": "%s"}`, custom_errors.InvalidJSONErr.Error())
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.JSONEq(t, expectedJSON, w.Body.String())
}

func TestGetStoresWrongSort(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := NewHandler(mockPool, "", "", 0)
	mux := NewStoresRouter(h)

	body, _ := json.Marshal(db.GetRequest{Limit: 10, Sorted: "id -- drop table"})
	req := httptest.NewRequest(http.MethodPost, "/stores", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	expectedJSON := fmt.Sprintf(`{"error": "%s"}`, custom_errors.InvalidJSONErr.Error())
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.JSONEq(t, expectedJSON, w.Body.String())
}
