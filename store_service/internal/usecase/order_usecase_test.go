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

	uid := "00000000-0000-0000-0000-000000000001"
	item := &domain.OrderItemInfo{
		ID:       uid,
		Name:     "item1",
		CardImg:  "card1",
		Price:    105.5,
		Quantity: 1,
	}

	order := &domain.OrderInfo{
		ID:        uid,
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
				id:  uid,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					GetOrder(context.Background(), uid).
					Return(order, nil)
			},
			expectedResult: order,
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
					GetOrder(context.Background(), uid).
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

			orders, err := uc.GetOrder(tt.input.ctx, tt.input.id)

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
		ctx    context.Context
		id     string
		status string
	}

	type testCase struct {
		name          string
		input         args
		mockSetup     func(repo *mock.MockOrderRepository)
		expectedError error
	}

	uid := "00000000-0000-0000-0000-000000000001"
	pending := "pending"
	paid := "paid"
	delivered := "delivered"
	cancelled := "cancelled"
	onTheWay := "on the way"

	tests := []testCase{
		{
			name: "успешный вызов pending",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: pending,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					UpdateOrderStatus(context.Background(), uid, pending).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов onTheWay",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: onTheWay,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					UpdateOrderStatus(context.Background(), uid, onTheWay).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов paid",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: paid,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					UpdateOrderStatus(context.Background(), uid, paid).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов cancelled",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: cancelled,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					UpdateOrderStatus(context.Background(), uid, cancelled).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "успешный вызов delivered",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: delivered,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					UpdateOrderStatus(context.Background(), uid, delivered).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "некорректный статус",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: "оплачено",
			},
			mockSetup:     func(repo *mock.MockOrderRepository) {},
			expectedError: domain.ErrRequestParams,
		},
		{
			name: "ошибка выполнения",
			input: args{
				ctx:    context.Background(),
				id:     uid,
				status: pending,
			},
			mockSetup: func(repo *mock.MockOrderRepository) {
				repo.EXPECT().
					UpdateOrderStatus(context.Background(), uid, pending).
					Return(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
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

			err := uc.UpdateOrderStatus(tt.input.ctx, tt.input.id, tt.input.status)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
