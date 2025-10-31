package usecase

import (
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/usecase/mock"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCartUsecase_GetCart(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name           string
		input          args
		mockSetup      func(repo *mock.MockCartRepository)
		expectedResult *domain.Cart
		expectedError  error
	}

	uid := "00000000-0000-0000-0000-000000000001"
	item := &domain.CartItem{
		ID:       uid,
		Name:     "name1",
		CardImg:  "img1",
		Price:    10.,
		Quantity: 1,
	}

	tests := []testCase{
		{
			name: "успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  uid,
			},
			mockSetup: func(repo *mock.MockCartRepository) {
				repo.EXPECT().
					GetCartItems(context.Background(), uid).
					Return([]*domain.CartItem{item}, nil)
			},
			expectedResult: &domain.Cart{
				Items: []*domain.CartItem{item},
			},
			expectedError: nil,
		},
		{
			name: "ошбика выполнения",
			input: args{
				ctx: context.Background(),
				id:  uid,
			},
			mockSetup: func(repo *mock.MockCartRepository) {
				repo.EXPECT().
					GetCartItems(context.Background(), uid).
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

			mockRepo := mock.NewMockCartRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := NewCartUsecase(mockRepo)

			cart, err := uc.GetCart(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, cart)
		})
	}
}

func TestCartUsecase_UpdateCart(t *testing.T) {
	type args struct {
		ctx      context.Context
		id       string
		updItems *domain.CartUpdate
	}

	type testCase struct {
		name           string
		input          args
		mockSetup      func(repo *mock.MockCartRepository)
		expectedResult *domain.Cart
		expectedError  error
	}

	uid := "00000000-0000-0000-0000-000000000001"
	item := &domain.ItemUpdate{
		ID:       uid,
		Quantity: 1,
	}
	cartUpd := &domain.CartUpdate{
		Items: []*domain.ItemUpdate{item},
	}

	tests := []testCase{
		{
			name: "успешный вызов",
			input: args{
				ctx:      context.Background(),
				id:       uid,
				updItems: cartUpd,
			},
			mockSetup: func(repo *mock.MockCartRepository) {
				repo.EXPECT().
					UpdateCartItems(context.Background(), uid, cartUpd).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "ошбика выполнения",
			input: args{
				ctx:      context.Background(),
				id:       uid,
				updItems: cartUpd,
			},
			mockSetup: func(repo *mock.MockCartRepository) {
				repo.EXPECT().
					UpdateCartItems(context.Background(), uid, cartUpd).
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

			mockRepo := mock.NewMockCartRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := NewCartUsecase(mockRepo)

			err := uc.UpdateCart(tt.input.ctx, tt.input.id, tt.input.updItems)

			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCartUsecase_DeleteCart(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name          string
		input         args
		mockSetup     func(repo *mock.MockCartRepository)
		expectedError error
	}

	uid := "00000000-0000-0000-0000-000000000001"

	tests := []testCase{
		{
			name: "успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  uid,
			},
			mockSetup: func(repo *mock.MockCartRepository) {
				repo.EXPECT().
					DeleteCartItems(context.Background(), uid).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "ошбика выполнения",
			input: args{
				ctx: context.Background(),
				id:  uid,
			},
			mockSetup: func(repo *mock.MockCartRepository) {
				repo.EXPECT().
					DeleteCartItems(context.Background(), uid).
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

			mockRepo := mock.NewMockCartRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := NewCartUsecase(mockRepo)

			err := uc.DeleteCart(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
