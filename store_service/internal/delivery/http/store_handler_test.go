package http

import (
	"apple_backend/handlers"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/mock"
	"apple_backend/store_service/internal/delivery/transport"
	"apple_backend/store_service/internal/domain"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestStoreHandler_GetStore(t *testing.T) {
	url := "/stores/%s"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockStoreUsecaseInterface)
		expectedCode      int
		expectedResult    *transport.StoreResponse
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"

	tests := []testCase{
		{
			name:   "GetStore успешный вызов",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStore(context.Background(), uid1).
					Return(&domain.StoreAgg{
						ID:          uid1,
						Name:        "name",
						Description: "description",
						CityID:      uid1,
						Address:     "address",
						CardImg:     "card_img",
						Rating:      5,
						OpenAt:      "open_at",
						TagsID:      []string{uid1},
						ClosedAt:    "closed_at",
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: &transport.StoreResponse{
				ID:          uid1,
				Name:        "name",
				Description: "description",
				CityID:      uid1,
				Address:     "address",
				CardImg:     "card_img",
				Rating:      5,
				TagsID:      []string{uid1},
				OpenAt:      "open_at",
				ClosedAt:    "closed_at",
			},
		},
		{
			name:              "GetStore метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "GetStore неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "GetStore не найдено данных",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStore(context.Background(), uid1).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetStore внутренняя ошибка",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStore(context.Background(), uid1).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockStoreUsecaseInterface(ctrl)
	handler := NewStoreHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.GetStore(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}

func TestStoreHandler_GetStores(t *testing.T) {
	url := "/stores"
	type testCase struct {
		name              string
		method            string
		body              string
		mockSetup         func(uc *mock.MockStoreUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.StoreResponse
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	name1 := "name1"
	name2 := "name2"
	description1 := "description1"
	description2 := "description2"
	address1 := "address1"
	address2 := "address2"
	cardImg1 := "card_img1"
	cardImg2 := "card_img2"
	rating1 := 1.
	rating2 := 2.
	openAt1 := "open_at1"
	openAt2 := "open_at2"
	closedAt1 := "closed_at1"
	closedAt2 := "closed_at2"

	store1 := &domain.StoreAgg{
		ID:          uid1,
		Name:        name1,
		Description: description1,
		CityID:      uid1,
		Address:     address1,
		CardImg:     cardImg1,
		Rating:      rating1,
		TagsID:      []string{uid1},
		OpenAt:      openAt1,
		ClosedAt:    closedAt1,
	}

	store2 := &domain.StoreAgg{
		ID:          uid2,
		Name:        name2,
		Description: description2,
		CityID:      uid2,
		Address:     address2,
		CardImg:     cardImg2,
		Rating:      rating2,
		TagsID:      []string{uid2},
		OpenAt:      openAt2,
		ClosedAt:    closedAt2,
	}

	storeResp1 := &transport.StoreResponse{
		ID:          store1.ID,
		Name:        store1.Name,
		Description: store1.Description,
		CityID:      store1.CityID,
		Address:     store1.Address,
		CardImg:     store1.CardImg,
		Rating:      store1.Rating,
		TagsID:      []string{uid1},
		OpenAt:      store1.OpenAt,
		ClosedAt:    store1.ClosedAt,
	}

	storeResp2 := &transport.StoreResponse{
		ID:          uid2,
		Name:        name2,
		Description: description2,
		CityID:      uid2,
		Address:     address2,
		CardImg:     cardImg2,
		Rating:      rating2,
		TagsID:      []string{uid2},
		OpenAt:      openAt2,
		ClosedAt:    closedAt2,
	}

	tests := []testCase{
		{
			name:   "GetStores успешный вызов без фильтров",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 10}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 10}).
					Return([]*domain.StoreAgg{store1, store2}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.StoreResponse{
				storeResp1,
				storeResp2,
			},
		},
		{
			name:   "GetStores успешный вызов с фильтром по тегу",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 10, TagID: uid1}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 10, TagID: uid1}).
					Return([]*domain.StoreAgg{store1}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.StoreResponse{
				storeResp1,
			},
		},
		{
			name:   "GetStores успешный вызов с фильтром по городу",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 10, CityID: uid2}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 10, CityID: uid2}).
					Return([]*domain.StoreAgg{store2}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.StoreResponse{
				storeResp2,
			},
		},
		{
			name:   "GetStores успешный вызов с сортировкой",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 5, Sorted: "rating"}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 5, Sorted: "rating"}).
					Return([]*domain.StoreAgg{store2, store1}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.StoreResponse{
				storeResp2,
				storeResp1,
			},
		},
		{
			name:              "GetStores метод не разрешен",
			method:            http.MethodGet,
			body:              parseJSON(&domain.StoreFilter{Limit: 10}),
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "GetStores неверный формат json",
			method:            http.MethodPost,
			body:              "запрос",
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "GetStores не найдено данных",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 10}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 10}).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetStores некорректные данные фильтра",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 0}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 0}).
					Return(nil, domain.ErrRequestParams)
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "GetStores внутренняя ошибка",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 10}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 10}).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockStoreUsecaseInterface(ctrl)
	handler := NewStoreHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, url, bytes.NewBuffer([]byte(tt.body)))
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()

			handler.GetStores(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}

func TestStoreHandler_CreateStore(t *testing.T) {
	url := "/stores"
	type testCase struct {
		name              string
		method            string
		body              string
		mockSetup         func(uc *mock.MockStoreUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.StoreResponse
		expectedErrResult *handlers.ErrResponse
	}

	store := &domain.Store{
		Name:        "name",
		Description: "description",
		CityID:      "00000000-0000-0000-0000-000000000001",
		Address:     "address",
		CardImg:     "card_img",
		Rating:      5,
		OpenAt:      "open_at",
		ClosedAt:    "closed_at",
	}

	tests := []testCase{
		{
			name:   "CreateStore успешный вызов",
			method: http.MethodPost,
			body:   parseJSON(store),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					CreateStore(context.Background(), store.Name, store.Description, store.CityID,
						store.Address, store.CardImg, store.OpenAt, store.ClosedAt, store.Rating).
					Return(nil)
			},
			expectedCode:   http.StatusCreated,
			expectedResult: nil,
		},
		{
			name:              "CreateStore метод не разрешен",
			method:            http.MethodPut,
			body:              parseJSON(store),
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "CreateStore неверный формат json",
			method:            http.MethodPost,
			body:              "запрос",
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "CreateStore уже существующий магазин",
			method: http.MethodPost,
			body:   parseJSON(store),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					CreateStore(context.Background(), store.Name, store.Description, store.CityID,
						store.Address, store.CardImg, store.OpenAt, store.ClosedAt, store.Rating).
					Return(domain.ErrStoreExist)
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrStoreExist.Error()},
		},
		{
			name:   "CreateStore внутренняя ошибка",
			method: http.MethodPost,
			body:   parseJSON(store),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					CreateStore(context.Background(), store.Name, store.Description, store.CityID,
						store.Address, store.CardImg, store.OpenAt, store.ClosedAt, store.Rating).
					Return(domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockStoreUsecaseInterface(ctrl)
	handler := NewStoreHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, url, bytes.NewBuffer([]byte(tt.body)))
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()

			handler.CreateStore(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}

func TestStoreHandler_GetStoreReview(t *testing.T) {
	url := "/stores/%s/reviews"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockStoreUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.StoreReview
		expectedErrResult *handlers.ErrResponse
	}

	storeID := "00000000-0000-0000-0000-000000000001"
	userName1 := "user1"
	userName2 := "user1"
	rating1 := 0.5
	rating2 := 5.
	comment1 := "comment1"
	comment2 := "comment2"
	createdAt1 := "2024-01-01"
	createdAt2 := "2024-02-01"

	reviw1 := &domain.StoreReview{
		UserName:  userName1,
		Rating:    rating1,
		Comment:   comment1,
		CreatedAt: createdAt1,
	}
	reviw2 := &domain.StoreReview{
		UserName:  userName2,
		Rating:    rating2,
		Comment:   comment2,
		CreatedAt: createdAt2,
	}

	reviewResp1 := &transport.StoreReview{
		UserName:  userName1,
		Rating:    rating1,
		Comment:   comment1,
		CreatedAt: createdAt1,
	}
	reviewResp2 := &transport.StoreReview{
		UserName:  userName2,
		Rating:    rating2,
		Comment:   comment2,
		CreatedAt: createdAt2,
	}

	tests := []testCase{
		{
			name:   "успешный вызов",
			method: http.MethodGet,
			id:     storeID,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStoreReview(context.Background(), storeID).
					Return([]*domain.StoreReview{reviw1, reviw2}, nil)
			},
			expectedCode:   http.StatusOK,
			expectedResult: []*transport.StoreReview{reviewResp1, reviewResp2},
		},
		{
			name:              "метод не разрешен",
			method:            http.MethodPost,
			id:                storeID,
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "GetStore неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "не найдено данных",
			method: http.MethodGet,
			id:     storeID,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStoreReview(context.Background(), storeID).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodGet,
			id:     storeID,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStoreReview(context.Background(), storeID).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mock.NewMockStoreUsecaseInterface(ctrl)
			handler := NewStoreHandler(uc, logger.NewNilLogger())
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.GetStoreReview(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}

func TestStoreHandler_GetCities(t *testing.T) {
	url := "/cities"
	type testCase struct {
		name              string
		method            string
		mockSetup         func(uc *mock.MockStoreUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.CityResponse
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	name1 := "name1"
	uid2 := "00000000-0000-0000-0000-000000000002"
	name2 := "name2"

	tests := []testCase{
		{
			name:   "GetCities успешный вызов",
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetCities(context.Background()).
					Return([]*domain.City{
						{
							ID:   uid1,
							Name: name1,
						},
						{
							ID:   uid2,
							Name: name2,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.CityResponse{
				{
					ID:   uid1,
					Name: name1,
				},
				{
					ID:   uid2,
					Name: name2,
				},
			},
		},
		{
			name:              "GetCities метод не разрешен",
			method:            http.MethodPost,
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:   "GetCities не найдено данных",
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetCities(context.Background()).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetCities внутренняя ошибка",
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetCities(context.Background()).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockStoreUsecaseInterface(ctrl)
	handler := NewStoreHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, url, bytes.NewBuffer([]byte(nil)))
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()

			handler.GetCities(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}

func TestStoreHandler_GetTags(t *testing.T) {
	url := "/tags"
	type testCase struct {
		name              string
		method            string
		mockSetup         func(uc *mock.MockStoreUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.TagResponse
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	name1 := "name1"
	uid2 := "00000000-0000-0000-0000-000000000002"
	name2 := "name2"

	tests := []testCase{
		{
			name:   "GetTags успешный вызов",
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetTags(context.Background()).
					Return([]*domain.StoreTag{
						{
							ID:   uid1,
							Name: name1,
						},
						{
							ID:   uid2,
							Name: name2,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.TagResponse{
				{
					ID:   uid1,
					Name: name1,
				},
				{
					ID:   uid2,
					Name: name2,
				},
			},
		},
		{
			name:              "GetTags метод не разрешен",
			method:            http.MethodPost,
			mockSetup:         func(uc *mock.MockStoreUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:   "GetTags не найдено данных",
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetTags(context.Background()).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetTags внутренняя ошибка",
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetTags(context.Background()).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockStoreUsecaseInterface(ctrl)
	handler := NewStoreHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, url, bytes.NewBuffer([]byte(nil)))
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()

			handler.GetTags(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}
