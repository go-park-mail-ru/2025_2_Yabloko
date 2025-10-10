package usecase

import (
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/usecase/mock"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestItemUsecase_GetItems(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name           string
		input          args
		expectedResult []*domain.Item
		expectedError  error
	}

	tests := []testCase{
		{
			name: "GetItems успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			expectedResult: []*domain.Item{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "name",
					Price:       9.99,
					Description: "description",
					CardImg:     "card_img",
					TypesID:     []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"},
				},
				{
					ID:          "00000000-0000-0000-0000-000000000002",
					Name:        "name",
					Price:       2.50,
					Description: "description",
					CardImg:     "card_img",
					TypesID:     []string{"00000000-0000-0000-0000-000000000003"},
				},
			},
			expectedError: nil,
		},
		{
			name: "GetItems ошибка выполнения",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			expectedResult: nil,
			expectedError:  errors.New("custom error"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockItemRepository(ctrl)
	uc := NewItemUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.EXPECT().
				GetItems(tt.input.ctx, tt.input.id).
				Return(tt.expectedResult, tt.expectedError)

			result, err := uc.GetItems(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestItemUsecase_GetItemTypes(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name           string
		input          args
		expectedResult []*domain.ItemType
		expectedError  error
	}

	tests := []testCase{
		{
			name: "GetItemTypes успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			expectedResult: []*domain.ItemType{
				{ID: "00000000-0000-0000-0000-000000000001", Name: "Пицца"},
				{ID: "00000000-0000-0000-0000-000000000002", Name: "Напитки"},
			},
			expectedError: nil,
		},
		{
			name: "GetItemTypes ошибка выполнения",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			expectedResult: nil,
			expectedError:  errors.New("custom error"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockItemRepository(ctrl)
	uc := NewItemUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.EXPECT().
				GetItemTypes(tt.input.ctx, tt.input.id).
				Return(tt.expectedResult, tt.expectedError)

			result, err := uc.GetItemTypes(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}
