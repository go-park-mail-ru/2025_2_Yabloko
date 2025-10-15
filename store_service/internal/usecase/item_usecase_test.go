package usecase

import (
	"apple_backend/custom_errors"
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
		repoResponse   []*domain.Item
		repoError      error
		expectedResult []*domain.ItemAgg
		expectedError  error
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	name1 := "name1"
	description1 := "description1"
	price1 := 1.0
	cardImg1 := "card_img1"

	uid2 := "00000000-0000-0000-0000-000000000002"
	name2 := "name2"
	description2 := "description2"
	price2 := 2.0
	cardImg2 := "card_img2"

	tests := []testCase{
		{
			name: "GetItems успешный вызов без необходимости аггрегаций",
			input: args{
				ctx: context.Background(),
				id:  uid1,
			},
			repoResponse: []*domain.Item{
				{
					ID:          uid1,
					Name:        name1,
					Price:       price1,
					Description: description1,
					CardImg:     cardImg1,
					TypeID:      uid1,
				},
				{
					ID:          uid2,
					Name:        name2,
					Price:       price2,
					Description: description2,
					CardImg:     cardImg2,
					TypeID:      uid2,
				},
			},
			repoError: nil,
			expectedResult: []*domain.ItemAgg{
				{
					ID:          uid1,
					Name:        name1,
					Price:       price1,
					Description: description1,
					CardImg:     cardImg1,
					TypesID:     []string{uid1},
				},
				{
					ID:          uid2,
					Name:        name2,
					Price:       price2,
					Description: description2,
					CardImg:     cardImg2,
					TypesID:     []string{uid2},
				},
			},
			expectedError: nil,
		},
		{
			name: "GetItems успешный вызов без необходимости аггрегаций",
			input: args{
				ctx: context.Background(),
				id:  uid1,
			},
			repoResponse: []*domain.Item{
				{
					ID:          uid1,
					Name:        name1,
					Price:       price1,
					Description: description1,
					CardImg:     cardImg1,
					TypeID:      uid1,
				},
				{
					ID:          uid2,
					Name:        name2,
					Price:       price2,
					Description: description2,
					CardImg:     cardImg2,
					TypeID:      uid2,
				},
				{
					ID:          uid1,
					Name:        name1,
					Price:       price1,
					Description: description1,
					CardImg:     cardImg1,
					TypeID:      uid2,
				},
				{
					ID:          uid2,
					Name:        name2,
					Price:       price2,
					Description: description2,
					CardImg:     cardImg2,
					TypeID:      uid1,
				},
			},
			repoError: nil,
			expectedResult: []*domain.ItemAgg{
				{
					ID:          uid1,
					Name:        name1,
					Price:       price1,
					Description: description1,
					CardImg:     cardImg1,
					TypesID:     []string{uid1, uid2},
				},
				{
					ID:          uid2,
					Name:        name2,
					Price:       price2,
					Description: description2,
					CardImg:     cardImg2,
					TypesID:     []string{uid2, uid1},
				},
			},
			expectedError: nil,
		},
		{
			name: "GetItems ошибка выполнения",
			input: args{
				ctx: context.Background(),
				id:  uid1,
			},
			repoResponse:   nil,
			repoError:      custom_errors.InnerErr,
			expectedResult: nil,
			expectedError:  custom_errors.InnerErr,
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
				Return(tt.repoResponse, tt.repoError)

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
