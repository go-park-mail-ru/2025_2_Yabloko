package usecase

import (
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/usecase/mock"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestOrderUsecase_GetOrdersUser(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name           string
		input          args
		mockSetup      func(repo *mock.MockOrderRepository)
		expectedResult []*domain.Order
		expectedError  error
	}

	uid := "00000000-0000-0000-0000-000000000001"
	order := &domain.Order{
		ID:        uid,
		Status:    "on the way",
		Total:     105.5,
		CreatedAt: "2024-01-01",
	}

	tests := []testCase{
		{
			name: "успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  uid,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					GetOrdersUser(context.Background(), uid).
					Return([]*domain.Order{order}, nil)
			},
			expectedResult: []*domain.Order{order},
			expectedError:  nil,
		},
		{
			name: "ошбика выполнения",
			input: args{
				ctx: context.Background(),
				id:  uid,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					GetOrdersUser(context.Background(), uid).
					Return(nil, domain.ErrInternalServer)
			},
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockOrderRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := NewOrderUsecase(mockRepo)

			orders, err := uc.GetOrdersUser(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, orders)
		})
	}
}

func TestOrderUsecase_GetOrder(t *testing.T) {
	type args struct {
		ctx     context.Context
		orderID string
		userID  string
	}

	type testCase struct {
		name           string
		input          args
		mockSetup      func(repo *mock.MockOrderRepository, orderID, userID string)
		expectedResult *domain.OrderInfo
		expectedError  error
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	item := &domain.OrderItemInfo{
		ID:       uid1,
		Name:     "item1",
		CardImg:  "card1",
		Price:    105.5,
		Quantity: 1,
	}

	order := &domain.OrderInfo{
		ID:        uid1,
		Items:     []*domain.OrderItemInfo{item},
		Status:    "on the way",
		Total:     105.5,
		CreatedAt: "2024-01-01",
	}

	tests := []testCase{
		{
			name: "успешный вызов",
			input: args{
				ctx:     context.Background(),
				userID:  uid1,
				orderID: uid2,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					GetOrder(context.Background(), orderID).
					Return(order, nil)
			},
			expectedResult: order,
			expectedError:  nil,
		},
		{
			name: "не тот пользователь",
			input: args{
				ctx:     context.Background(),
				userID:  uid1,
				orderID: uid2,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(orderID, nil)
			},
			expectedResult: nil,
			expectedError:  domain.ErrForbidden,
		},
		{
			name: "ошибка при получении пользователя",
			input: args{
				ctx:     context.Background(),
				userID:  uid1,
				orderID: uid2,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return("", domain.ErrInternalServer)
			},
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
		{
			name: "ошибка получения заказа",
			input: args{
				ctx:     context.Background(),
				userID:  uid1,
				orderID: uid2,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					GetOrder(context.Background(), orderID).
					Return(nil, domain.ErrInternalServer)
			},
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockOrderRepository(ctrl)
			tt.mockSetup(mockRepo, tt.input.orderID, tt.input.userID)

			uc := NewOrderUsecase(mockRepo)

			orders, err := uc.GetOrder(tt.input.ctx, tt.input.orderID, tt.input.userID)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, orders)
		})
	}
}

func TestOrderUsecase_CreateOrder(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name           string
		input          args
		mockSetup      func(repo *mock.MockOrderRepository)
		expectedResult *domain.OrderInfo
		expectedError  error
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	item := &domain.OrderItemInfo{
		ID:       uid1,
		Name:     "item1",
		CardImg:  "card1",
		Price:    105.5,
		Quantity: 1,
	}

	order := &domain.OrderInfo{
		ID:        uid2,
		Items:     []*domain.OrderItemInfo{item},
		Status:    "on the way",
		Total:     105.5,
		CreatedAt: "2024-01-01",
	}

	tests := []testCase{
		{
			name: "успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  uid1,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					CreateOrder(context.Background(), uid1).
					Return(uid2, nil)
				repo.EXPECT().
					GetOrder(context.Background(), uid2).
					Return(order, nil)
			},
			expectedResult: order,
			expectedError:  nil,
		},
		{
			name: "ошбика выполнения после создания заказа",
			input: args{
				ctx: context.Background(),
				id:  uid1,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					CreateOrder(context.Background(), uid1).
					Return("", domain.ErrInternalServer)
			},
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
		{
			name: "ошбика выполнения после получения заказа",
			input: args{
				ctx: context.Background(),
				id:  uid1,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					CreateOrder(context.Background(), uid1).
					Return(uid2, nil)
				repo.EXPECT().
					GetOrder(context.Background(), uid2).
					Return(nil, domain.ErrInternalServer)
			},
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockOrderRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := NewOrderUsecase(mockRepo)

			orders, err := uc.CreateOrder(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, orders)
		})
	}
}

func TestOrderUsecase_UpdateOrderStatus(t *testing.T) {
	type args struct {
		ctx     context.Context
		orderID string
		userID  string
		status  string
	}

	type testCase struct {
		name          string
		input         args
		mockSetup     func(repo *mock.MockOrderRepository, orderID, userID string)
		expectedError error
	}

	uid := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	pending := "pending"
	paid := "paid"
	delivered := "delivered"
	cancelled := "cancelled"
	onTheWay := "on the way"

	tests := []testCase{
		{
			name: "успешный вызов pending",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  pending,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					UpdateOrderStatus(context.Background(), orderID, pending).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов paid",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  paid,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					UpdateOrderStatus(context.Background(), orderID, paid).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов delivered",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  delivered,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					UpdateOrderStatus(context.Background(), orderID, delivered).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов cancelled",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  cancelled,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					UpdateOrderStatus(context.Background(), orderID, cancelled).
					Return(nil)
			},
			expectedError: nil,
		},

		{
			name: "успешный вызов onTheWay",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  onTheWay,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					UpdateOrderStatus(context.Background(), orderID, onTheWay).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "некорректный статус",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  "оплачено",
			},
			mockSetup:     func(repo *mock.MockOrderRepository, orderID, userID string) {},
			expectedError: domain.ErrRequestParams,
		},
		{
			name: "ошибка получения ид",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  pending,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return("", domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "ошибка обновления статуса",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  pending,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(userID, nil)

				repo.EXPECT().
					UpdateOrderStatus(context.Background(), orderID, pending).
					Return(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "некорректный userID",
			input: args{
				ctx:     context.Background(),
				orderID: uid,
				userID:  uid2,
				status:  pending,
			},
			mockSetup: func(repo *mock.MockOrderRepository, orderID, userID string) {
				repo.EXPECT().
					GetOrderUserID(context.Background(), orderID).
					Return(orderID, nil)
			},
			expectedError: domain.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockOrderRepository(ctrl)
			tt.mockSetup(mockRepo, tt.input.orderID, tt.input.userID)

			uc := NewOrderUsecase(mockRepo)

			err := uc.UpdateOrderStatus(tt.input.ctx, tt.input.orderID, tt.input.userID, tt.input.status)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
