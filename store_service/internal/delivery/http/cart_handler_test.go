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

func TestCartHandler_GetCart(t *testing.T) {
	url := "/users/%s/cart"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockCartUsecaseInterface)
		expectedCode      int
		expectedResult    *transport.Cart
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	name1 := "name1"
	name2 := "name2"
	cardImg1 := "cardImg1"
	cardImg2 := "cardImg2"
	price1 := 1.
	price2 := 2.
	quantity1 := 12
	quantity2 := 13

	item1 := &transport.CartItem{
		ID:       uid1,
		Name:     name1,
		Price:    price1,
		Quantity: quantity1,
		CardImg:  cardImg1,
	}
	item2 := &transport.CartItem{
		ID:       uid2,
		Name:     name2,
		Price:    price2,
		Quantity: quantity2,
		CardImg:  cardImg2,
	}

	itemUC1 := &domain.CartItem{
		ID:       uid1,
		Name:     name1,
		Price:    price1,
		Quantity: quantity1,
		CardImg:  cardImg1,
	}
	itemUC2 := &domain.CartItem{
		ID:       uid2,
		Name:     name2,
		Price:    price2,
		Quantity: quantity2,
		CardImg:  cardImg2,
	}

	tests := []testCase{
		{
			name:   "успешный вызов",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockCartUsecaseInterface) {
				uc.EXPECT().
					GetCart(context.Background(), uid1).
					Return(&domain.Cart{
						ID: uid1,
						Items: []*domain.CartItem{
							itemUC1,
							itemUC2,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: &transport.Cart{
				ID: uid1,
				Items: []*transport.CartItem{
					item1,
					item2,
				},
			},
			expectedErrResult: nil,
		},
		{
			name:              "метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			mockSetup:         func(uc *mock.MockCartUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockCartUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "не найдено данных",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockCartUsecaseInterface) {
				uc.EXPECT().
					GetCart(context.Background(), uid1).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetItemTypes внутренняя ошибка",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockCartUsecaseInterface) {
				uc.EXPECT().
					GetCart(context.Background(), uid1).
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

			uc := mock.NewMockCartUsecaseInterface(ctrl)
			handler := NewCartHandler(uc, logger.NewNilLogger())

			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.GetCart(w, req)

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

func TestCartHandler_UpdateCart(t *testing.T) {
	url := "/carts/%s"
	type testCase struct {
		name              string
		method            string
		id                string
		body              string
		mockSetup         func(uc *mock.MockCartUsecaseInterface)
		expectedCode      int
		expectedResult    *transport.UpdateResponse
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	quantity1 := 12
	quantity2 := 13

	item1 := &transport.ItemUpdate{
		ID:       uid1,
		Quantity: quantity1,
	}
	item2 := &transport.ItemUpdate{
		ID:       uid2,
		Quantity: quantity2,
	}

	itemUC1 := &domain.ItemUpdate{
		ID:       uid1,
		Quantity: quantity1,
	}
	itemUC2 := &domain.ItemUpdate{
		ID:       uid2,
		Quantity: quantity2,
	}

	tests := []testCase{
		{
			name:   "успешный вызов",
			method: http.MethodPut,
			id:     uid1,
			body:   parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup: func(uc *mock.MockCartUsecaseInterface) {
				uc.EXPECT().
					UpdateCart(context.Background(), uid1, &domain.CartUpdate{Items: []*domain.ItemUpdate{itemUC1, itemUC2}}).
					Return(nil)
			},
			expectedCode:      http.StatusOK,
			expectedResult:    &transport.UpdateResponse{ID: uid1},
			expectedErrResult: nil,
		},
		{
			name:              "метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			body:              parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup:         func(uc *mock.MockCartUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "неверный формат id",
			method:            http.MethodPut,
			id:                "00000000-1",
			body:              parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup:         func(uc *mock.MockCartUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:              "неверное тело запроса",
			method:            http.MethodPut,
			id:                uid1,
			body:              "{\"ID\": \"010210\"}",
			mockSetup:         func(uc *mock.MockCartUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:              "некорректный json",
			method:            http.MethodPut,
			id:                uid1,
			body:              "{\"ID\" \"010210\"}",
			mockSetup:         func(uc *mock.MockCartUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodPut,
			id:     uid1,
			body:   parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup: func(uc *mock.MockCartUsecaseInterface) {
				uc.EXPECT().
					UpdateCart(context.Background(), uid1, &domain.CartUpdate{Items: []*domain.ItemUpdate{itemUC1, itemUC2}}).
					Return(domain.ErrInternalServer)
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

			uc := mock.NewMockCartUsecaseInterface(ctrl)
			handler := NewCartHandler(uc, logger.NewNilLogger())

			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer([]byte(tt.body)))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.UpdateCart(w, req)

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
