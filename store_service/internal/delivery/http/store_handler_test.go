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
					Return(&domain.Store{
						ID:          uid1,
						Name:        "name",
						Description: "description",
						CityID:      uid1,
						Address:     "address",
						CardImg:     "card_img",
						Rating:      5,
						OpenAt:      "open_at",
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

	tests := []testCase{
		{
			name:   "GetStores успешный вызов",
			method: http.MethodPost,
			body:   parseJSON(&domain.StoreFilter{Limit: 10}),
			mockSetup: func(uc *mock.MockStoreUsecaseInterface) {
				uc.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 10}).
					Return([]*domain.Store{
						{
							ID:          uid1,
							Name:        "name",
							Description: "description",
							CityID:      uid1,
							Address:     "address",
							CardImg:     "card_img",
							Rating:      5,
							OpenAt:      "open_at",
							ClosedAt:    "closed_at",
						},
						{
							ID:          uid2,
							Name:        "name",
							Description: "description",
							CityID:      uid2,
							Address:     "address",
							CardImg:     "card_img",
							Rating:      5,
							OpenAt:      "open_at",
							ClosedAt:    "closed_at",
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.StoreResponse{
				{
					ID:          uid1,
					Name:        "name",
					Description: "description",
					CityID:      uid1,
					Address:     "address",
					CardImg:     "card_img",
					Rating:      5,
					OpenAt:      "open_at",
					ClosedAt:    "closed_at",
				},
				{
					ID:          uid2,
					Name:        "name",
					Description: "description",
					CityID:      uid2,
					Address:     "address",
					CardImg:     "card_img",
					Rating:      5,
					OpenAt:      "open_at",
					ClosedAt:    "closed_at",
				},
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
