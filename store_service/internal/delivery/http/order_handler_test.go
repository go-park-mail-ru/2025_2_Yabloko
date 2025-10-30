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

func TestOrderHandler_CreateOrder(t *testing.T) {
	url := "/users/%s/order"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface)
		expectedCode      int
		expectedResult    *transport.OrderInfo
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
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					CreateOrder(context.Background(), uid1).
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
			name:              "метод не разрешен",
			method:            http.MethodPut,
			id:                uid1,
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "неверный формат id",
			method:            http.MethodPost,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodPost,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					CreateOrder(context.Background(), uid1).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

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
		id                string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface)
		expectedCode      int
		expectedResult    *transport.OrderInfo
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
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					GetOrder(context.Background(), uid1).
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
			name:              "метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "не найдено заказа",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					GetOrder(context.Background(), uid1).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					GetOrder(context.Background(), uid1).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

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
	url := "/users/%s/orders"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface)
		expectedCode      int
		expectedResult    *transport.Orders
		expectedErrResult *handlers.ErrResponse
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
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					GetOrdersUser(context.Background(), uid).
					Return([]*domain.Order{orderUC}, nil)
			},
			expectedCode:      http.StatusOK,
			expectedResult:    &transport.Orders{Orders: []*transport.Order{order}},
			expectedErrResult: nil,
		},
		{
			name:              "метод не разрешен",
			method:            http.MethodPost,
			id:                uid,
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "не найдено заказов",
			method: http.MethodGet,
			id:     uid,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					GetOrdersUser(context.Background(), uid).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodGet,
			id:     uid,
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					GetOrdersUser(context.Background(), uid).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

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
		id                string
		body              string
		mockSetup         func(uc *mock.MockOrderUsecaseInterface)
		expectedCode      int
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	status := &transport.OrderStatus{Status: "on the way"}

	tests := []testCase{
		{
			name:   "успешный вызов",
			method: http.MethodPatch,
			id:     uid1,
			body:   parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					UpdateOrderStatus(context.Background(), uid1, status.Status).
					Return(nil)
			},
			expectedCode:      http.StatusNoContent,
			expectedErrResult: nil,
		},
		{
			name:              "метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			body:              parseJSON(status),
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "неверный формат id",
			method:            http.MethodPatch,
			id:                "00000000-1",
			body:              parseJSON(status),
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "не найдено заказа",
			method: http.MethodPatch,
			id:     uid1,
			body:   parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					UpdateOrderStatus(context.Background(), uid1, status.Status).
					Return(domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "некорректный статус",
			method: http.MethodPatch,
			id:     uid1,
			body:   "{\"status\": \"on\"}",
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					UpdateOrderStatus(context.Background(), uid1, "on").
					Return(domain.ErrRequestParams)
			},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:              "некорректное тело запроса",
			method:            http.MethodPatch,
			id:                uid1,
			body:              "{\"stat\": \"on\"}",
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:              "некорректный json",
			method:            http.MethodPatch,
			id:                uid1,
			body:              "{\"status\" \"on\"}",
			mockSetup:         func(uc *mock.MockOrderUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "внутренняя ошибка",
			method: http.MethodPatch,
			id:     uid1,
			body:   parseJSON(status),
			mockSetup: func(uc *mock.MockOrderUsecaseInterface) {
				uc.EXPECT().
					UpdateOrderStatus(context.Background(), uid1, status.Status).
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

			uc := mock.NewMockOrderUsecaseInterface(ctrl)
			handler := NewOrderHandler(uc, logger.NewNilLogger())

			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer([]byte(tt.body)))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.UpdateOrderStatus(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}
