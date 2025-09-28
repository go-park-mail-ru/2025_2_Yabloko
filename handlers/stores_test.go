package handlers

import (
	"apple_backend/custom_errors"
	"apple_backend/db"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestGetStoresLimit(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := New(mockPool, "", "", 0)
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", h.GetStores)

	storesData := [][]any{
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
			"Суши Мастер",
			"Современный японский ресторан с широким выбором суши и роллов.",
			uuid.MustParse("00000000-0000-0000-0000-000000000011").String(),
			"ул. Тверская, 12",
			"/stores/1.jpg",
			4.5,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			"Sushi House",
			"Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			uuid.MustParse("00000000-0000-0000-0000-000000000012").String(),
			"Невский пр., 45",
			"/stores/2.jpg",
			4.3,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000003").String(),
			"Pizza Roma",
			"Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			uuid.MustParse("00000000-0000-0000-0000-000000000013").String(),
			"ул. Ленина, 78",
			"/stores/3.jpg",
			4.6,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
	}

	expected := []db.ResponseInfo{
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
			"Суши Мастер",
			"Современный японский ресторан с широким выбором суши и роллов.",
			uuid.MustParse("00000000-0000-0000-0000-000000000011").String(),
			"ул. Тверская, 12",
			"/stores/1.jpg",
			4.5,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			"Sushi House",
			"Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			uuid.MustParse("00000000-0000-0000-0000-000000000012").String(),
			"Невский пр., 45",
			"/stores/2.jpg",
			4.3,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000003").String(),
			"Pizza Roma",
			"Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			uuid.MustParse("00000000-0000-0000-0000-000000000013").String(),
			"ул. Ленина, 78",
			"/stores/3.jpg",
			4.6,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
	}

	limits := []int{1, 2, 3}
	for _, limit := range limits {
		t.Run(fmt.Sprintf("Limit=%d", limit), func(t *testing.T) {
			rows := pgxmock.NewRows([]string{
				"id", "name", "description", "city_id", "address", "card_img",
				"rating", "open_at", "closed_at",
			})
			for _, store := range storesData[:limit] {
				rows.AddRow(store...)
			}

			query := `
				select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
				from store
				order by id
				limit \$1
			`
			mockPool.ExpectQuery(query).
				WithArgs(limit).
				WillReturnRows(rows)

			body, _ := json.Marshal(db.GetRequest{Limit: limit})
			req := httptest.NewRequest(http.MethodPost, "/stores", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			expectedJSON, _ := json.Marshal(expected[:limit])
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, string(expectedJSON), w.Body.String())
		})
	}
}

func TestGetStoresLimitWithID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := New(mockPool, "", "", 0)
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", h.GetStores)

	storesData := [][]any{
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
			"Суши Мастер",
			"Современный японский ресторан с широким выбором суши и роллов.",
			uuid.MustParse("00000000-0000-0000-0000-000000000011").String(),
			"ул. Тверская, 12",
			"/stores/1.jpg",
			4.5,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			"Sushi House",
			"Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			uuid.MustParse("00000000-0000-0000-0000-000000000012").String(),
			"Невский пр., 45",
			"/stores/2.jpg",
			4.3,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000003").String(),
			"Pizza Roma",
			"Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			uuid.MustParse("00000000-0000-0000-0000-000000000013").String(),
			"ул. Ленина, 78",
			"/stores/3.jpg",
			4.6,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
	}

	expected := []db.ResponseInfo{
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			"Sushi House",
			"Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
			uuid.MustParse("00000000-0000-0000-0000-000000000012").String(),
			"Невский пр., 45",
			"/stores/2.jpg",
			4.3,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
		{
			uuid.MustParse("00000000-0000-0000-0000-000000000003").String(),
			"Pizza Roma",
			"Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
			uuid.MustParse("00000000-0000-0000-0000-000000000013").String(),
			"ул. Ленина, 78",
			"/stores/3.jpg",
			4.6,
			time.Date(0, 1, 1, 10, 0, 0, 0, time.Local).String(),
			time.Date(0, 1, 1, 23, 0, 0, 0, time.Local).String(),
		},
	}
	requestBody := db.GetRequest{
		Limit:  3,
		LastId: "00000000-0000-0000-0000-000000000001"}

	rows := pgxmock.NewRows([]string{
		"id", "name", "description", "city_id", "address", "card_img",
		"rating", "open_at", "closed_at",
	})
	for _, store := range storesData[1:] {
		rows.AddRow(store...)
	}

	query := `
				select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
				from store
				where id > \$1
				order by id
				limit \$2
			`
	mockPool.ExpectQuery(query).
		WithArgs(requestBody.LastId, requestBody.Limit).
		WillReturnRows(rows)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/stores", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	expectedJSON, _ := json.Marshal(expected)
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, string(expectedJSON), w.Body.String())

}

func TestGetStoresEmptyDB(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := New(mockPool, "", "", 0)
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", h.GetStores)

	query := `
				select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
				from store
				order by id
				limit \$1
			`
	mockPool.ExpectQuery(query).
		WithArgs(3).
		WillReturnRows(pgxmock.NewRows([]string{}))

	body, _ := json.Marshal(db.GetRequest{Limit: 3})
	req := httptest.NewRequest(http.MethodPost, "/stores", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, "[]", w.Body.String())
}

func TestGetStoresNegativeLimit(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	h := New(mockPool, "", "", 0)
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", h.GetStores)

	body, _ := json.Marshal(db.GetRequest{Limit: -1})
	req := httptest.NewRequest(http.MethodPost, "/stores", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	expectedJSON := fmt.Sprintf(`{"error": "%s"}`, custom_errors.InvalidJSONErr.Error())
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.JSONEq(t, expectedJSON, w.Body.String())
}
