package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/middlewares"
	"apple_backend/store_service/internal/delivery/mock"
	"apple_backend/store_service/internal/delivery/transport"
	"apple_backend/store_service/internal/domain"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCartHandler_GetCart(t *testing.T) {
	url := "/cart"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockCartUsecaseInterface, method, userID string) *http.Request
		expectedCode      int
		expectedResult    *transport.Cart
		expectedErrResult *http_response.ErrResponse
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
			id:     uid1,
			method: http.MethodGet,
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetCart(ctx, userID).
					Return(&domain.Cart{
						Items: []*domain.CartItem{
							itemUC1,
							itemUC2,
						},
					}, nil)

				return req
			},
			expectedCode: http.StatusOK,
			expectedResult: &transport.Cart{
				Items: []*transport.CartItem{
					item1,
					item2,
				},
			},
			expectedErrResult: nil,
		},
		{
			name:   "метод не разрешен",
			method: http.MethodPost,
			id:     uid1,
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:   "неверный формат id",
			method: http.MethodGet,
			id:     "00000000-1",
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:   "не найдено данных",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetCart(ctx, userID).
					Return(nil, domain.ErrRowsNotFound)

				return req
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetCart(ctx, userID).
					Return(nil, domain.ErrInternalServer)

				return req
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mock.NewMockCartUsecaseInterface(ctrl)
			handler := NewCartHandler(uc, logger.NewNilLogger())

			req := tt.mockSetup(uc, tt.method, tt.id)
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
	url := "/carts"
	type testCase struct {
		name              string
		method            string
		id                string
		body              string
		mockSetup         func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request
		expectedCode      int
		expectedErrResult *http_response.ErrResponse
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
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					UpdateCart(ctx, userID, &domain.CartUpdate{Items: []*domain.ItemUpdate{itemUC1, itemUC2}}).
					Return(nil)

				return req
			},
			expectedCode:      http.StatusNoContent,
			expectedErrResult: nil,
		},
		{
			name:   "метод не разрешен",
			method: http.MethodPost,
			id:     uid1,
			body:   parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:   "неверный формат id",
			method: http.MethodPut,
			id:     "00000000-1",
			body:   parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:   "неверное тело запроса",
			method: http.MethodPut,
			id:     uid1,
			body:   "{\"ID\": \"010210\"}",
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "некорректный json",
			method: http.MethodPut,
			id:     uid1,
			body:   "{\"ID\" \"010210\"}",
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodPut,
			id:     uid1,
			body:   parseJSON(&transport.CartUpdate{Items: []*transport.ItemUpdate{item1, item2}}),
			mockSetup: func(uc *mock.MockCartUsecaseInterface, method, userID, body string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					UpdateCart(ctx, userID, &domain.CartUpdate{Items: []*domain.ItemUpdate{itemUC1, itemUC2}}).
					Return(domain.ErrInternalServer)

				return req
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mock.NewMockCartUsecaseInterface(ctrl)
			handler := NewCartHandler(uc, logger.NewNilLogger())

			req := tt.mockSetup(uc, tt.method, tt.id, tt.body)

			w := httptest.NewRecorder()

			handler.UpdateCart(w, req)

			require.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}
