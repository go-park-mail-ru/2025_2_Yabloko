package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/order_service/internal/delivery/middlewares"
	"apple_backend/order_service/internal/delivery/mock"
	"apple_backend/order_service/internal/delivery/transport"
	"apple_backend/order_service/internal/domain"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	url := "/orders"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request
		expectedCode      int
		expectedResult    *transport.OrderInfo
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

	status := "pending"
	total := price1 + price2
	createdAt := "2024-01-01"

	item1 := &transport.OrderItemInfo{
		ID:       uid1,
		Name:     name1,
		Price:    price1,
		Quantity: quantity1,
		CardImg:  cardImg1,
	}
	item2 := &transport.OrderItemInfo{
		ID:       uid2,
		Name:     name2,
		Price:    price2,
		Quantity: quantity2,
		CardImg:  cardImg2,
	}

	itemUC1 := &domain.OrderItemInfo{
		ID:       uid1,
		Name:     name1,
		Price:    price1,
		Quantity: quantity1,
		CardImg:  cardImg1,
	}
	itemUC2 := &domain.OrderItemInfo{
		ID:       uid2,
		Name:     name2,
		Price:    price2,
		Quantity: quantity2,
		CardImg:  cardImg2,
	}

	tests := []testCase{
		{
			name:   "успешный вызов",
			method: http.MethodPost,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					CreateOrder(ctx, userID).
					Return(&domain.OrderInfo{
						ID: uid1,
						Items: []*domain.OrderItemInfo{
							itemUC1,
							itemUC2,
						},
						Status:    status,
						Total:     total,
						CreatedAt: createdAt,
					}, nil)

				return req
			},
			expectedCode: http.StatusOK,
			expectedResult: &transport.OrderInfo{
				ID: uid1,
				Items: []*transport.OrderItemInfo{
					item1,
					item2,
				},
				Status:    status,
				Total:     total,
				CreatedAt: createdAt,
			},
			expectedErrResult: nil,
		},
		{
			name:   "метод не разрешен",
			method: http.MethodPut,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:   "не аутентифицирован",
			method: http.MethodPost,
			id:     "",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:   "неверный формат id",
			method: http.MethodPost,
			id:     "00000000-1",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodPost,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					CreateOrder(ctx, userID).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			req := tt.mockSetup(uc, tt.method, tt.id)

			w := httptest.NewRecorder()

			handler.CreateOrder(w, req)

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

func TestOrderHandler_GetOrder(t *testing.T) {
	url := "/orders/%s"
	type testCase struct {
		name              string
		method            string
		userID            string
		orderID           string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request
		expectedCode      int
		expectedResult    *transport.OrderInfo
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

	status := "pending"
	total := price1 + price2
	createdAt := "2024-01-01"

	item1 := &transport.OrderItemInfo{
		ID:       uid1,
		Name:     name1,
		Price:    price1,
		Quantity: quantity1,
		CardImg:  cardImg1,
	}
	item2 := &transport.OrderItemInfo{
		ID:       uid2,
		Name:     name2,
		Price:    price2,
		Quantity: quantity2,
		CardImg:  cardImg2,
	}

	itemUC1 := &domain.OrderItemInfo{
		ID:       uid1,
		Name:     name1,
		Price:    price1,
		Quantity: quantity1,
		CardImg:  cardImg1,
	}
	itemUC2 := &domain.OrderItemInfo{
		ID:       uid2,
		Name:     name2,
		Price:    price2,
		Quantity: quantity2,
		CardImg:  cardImg2,
	}

	tests := []testCase{
		{
			name:    "успешный вызов",
			method:  http.MethodGet,
			userID:  uid1,
			orderID: uid2,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrder(ctx, orderID, userID).
					Return(&domain.OrderInfo{
						ID: uid1,
						Items: []*domain.OrderItemInfo{
							itemUC1,
							itemUC2,
						},
						Status:    status,
						Total:     total,
						CreatedAt: createdAt,
					}, nil)

				return req
			},
			expectedCode: http.StatusOK,
			expectedResult: &transport.OrderInfo{
				ID: uid1,
				Items: []*transport.OrderItemInfo{
					item1,
					item2,
				},
				Status:    status,
				Total:     total,
				CreatedAt: createdAt,
			},
			expectedErrResult: nil,
		},
		{
			name:    "ошибка доступа",
			method:  http.MethodGet,
			userID:  uid1,
			orderID: uid2,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrder(ctx, orderID, userID).
					Return(nil, domain.ErrForbidden)

				return req
			},
			expectedCode:      http.StatusForbidden,
			expectedResult:    nil,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrForbidden.Error()},
		},
		{
			name:    "метод не разрешен",
			method:  http.MethodPost,
			userID:  uid1,
			orderID: uid2,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:    "неверный формат userID",
			method:  http.MethodGet,
			userID:  "00000000-1",
			orderID: uid2,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:    "неверный формат orderID",
			method:  http.MethodGet,
			userID:  uid1,
			orderID: "00000000-1",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:    "не найдено заказа",
			method:  http.MethodGet,
			userID:  uid1,
			orderID: uid2,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrder(ctx, orderID, userID).
					Return(nil, domain.ErrRowsNotFound)

				return req
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:    "внутренняя ошибка",
			method:  http.MethodGet,
			userID:  uid1,
			orderID: uid2,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, orderID, userID string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrder(ctx, orderID, userID).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			req := tt.mockSetup(uc, tt.method, tt.orderID, tt.userID)
			req.SetPathValue("id", tt.orderID)

			w := httptest.NewRecorder()

			handler.GetOrder(w, req)

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

func TestOrderHandler_GetOrdersUser(t *testing.T) {
	url := "/orders"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request
		expectedCode      int
		expectedResult    *transport.Orders
		expectedErrResult *http_response.ErrResponse
	}

	uid := "00000000-0000-0000-0000-000000000001"
	status := "on the way"
	total := 101.5
	createdAt := "2024-01-01"

	order := &transport.Order{
		ID:        uid,
		Status:    status,
		Total:     total,
		CreatedAt: createdAt,
	}
	orderUC := &domain.Order{
		ID:        uid,
		Status:    status,
		Total:     total,
		CreatedAt: createdAt,
	}

	tests := []testCase{
		{
			name:   "успешный вызов",
			method: http.MethodGet,
			id:     uid,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrdersUser(ctx, uid).
					Return([]*domain.Order{orderUC}, nil)

				return req
			},
			expectedCode:      http.StatusOK,
			expectedResult:    &transport.Orders{Orders: []*transport.Order{order}},
			expectedErrResult: nil,
		},
		{
			name:   "метод не разрешен",
			method: http.MethodPost,
			id:     uid,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
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
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:   "не найдено заказов",
			method: http.MethodGet,
			id:     uid,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrdersUser(ctx, uid).
					Return(nil, domain.ErrRowsNotFound)

				return req
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodGet,
			id:     uid,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					GetOrdersUser(ctx, uid).
					Return(nil, domain.ErrInternalServer)

				return req
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
		{
			name:   "нет авторизации",
			method: http.MethodGet,
			id:     "",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID string) *http.Request {
				req := httptest.NewRequest(method, url, bytes.NewBuffer(nil))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			req := tt.mockSetup(uc, tt.method, tt.id)

			w := httptest.NewRecorder()

			handler.GetOrdersUser(w, req)

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

func TestOrderHandler_UpdateOrderStatus(t *testing.T) {
	url := "/orders/%s/status"
	type testCase struct {
		name              string
		method            string
		userID            string
		orderID           string
		body              string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request
		expectedCode      int
		expectedErrResult *http_response.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	status := &transport.OrderStatus{Status: "on the way"}

	tests := []testCase{
		{
			name:    "успешный вызов",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: uid1,
			body:    parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					UpdateOrderStatus(ctx, orderID, userID, status.Status).
					Return(nil)

				return req
			},
			expectedCode:      http.StatusNoContent,
			expectedErrResult: nil,
		},
		{
			name:    "метод не разрешен",
			method:  http.MethodPost,
			userID:  uid1,
			orderID: uid1,
			body:    parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:    "неверный формат user_id",
			method:  http.MethodPatch,
			userID:  "00000000-1",
			orderID: uid1,
			body:    parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusUnauthorized,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrUnauthorized.Error()},
		},
		{
			name:    "неверный формат order_id",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: "00000000-1",
			body:    parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:    "не найдено заказа",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: uid1,
			body:    parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					UpdateOrderStatus(ctx, orderID, userID, status.Status).
					Return(domain.ErrRowsNotFound)

				return req
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:    "некорректный статус",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: uid1,
			body:    "{\"status\": \"on\"}",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					UpdateOrderStatus(ctx, orderID, userID, "on").
					Return(domain.ErrRequestParams)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:    "некорректное тело запроса",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: uid1,
			body:    "{\"stat\": \"on\"}",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:    "некорректный json",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: uid1,
			body:    "{\"status\" \"on\"}",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				return req
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &http_response.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:    "внутренняя ошибка",
			method:  http.MethodPatch,
			userID:  uid1,
			orderID: uid1,
			body:    parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface, method, userID, orderID, body string) *http.Request {
				req := httptest.NewRequest(method, fmt.Sprintf(url, orderID), bytes.NewBuffer([]byte(body)))
				ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
				req = req.WithContext(ctx)

				uc.EXPECT().
					UpdateOrderStatus(ctx, orderID, userID, status.Status).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			req := tt.mockSetup(uc, tt.method, tt.userID, tt.orderID, tt.body)
			req.SetPathValue("id", tt.orderID)

			w := httptest.NewRecorder()

			handler.UpdateOrderStatus(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}
