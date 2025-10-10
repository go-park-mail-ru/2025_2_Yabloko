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

func TestStoreUsecase_GetStore(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	type testCase struct {
		name           string
		input          args
		expectedResult *domain.Store
		expectedError  error
	}

	tests := []testCase{
		{
			name: "GetStore успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			expectedResult: &domain.Store{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "Store",
				Description: "Description",
				CityID:      "10000000-0000-0000-0000-000000000001",
				Address:     "Address",
				CardImg:     "CardImg",
				Rating:      3,
				OpenAt:      "OpenAt",
				ClosedAt:    "ClosedAt",
			},
			expectedError: nil,
		},
		{
			name: "GetStore ошбика выполнения",
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

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.EXPECT().
				GetStore(tt.input.ctx, tt.input.id).
				Return(tt.expectedResult, tt.expectedError)

			store, err := uc.GetStore(tt.input.ctx, tt.input.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, store)
		})
	}
}

func TestStoreUsecase_GetStores(t *testing.T) {
	type args struct {
		ctx    context.Context
		filter *domain.StoreFilter
	}

	type testCase struct {
		name           string
		input          args
		expectedResult []*domain.Store
		expectedError  error
	}

	tests := []testCase{
		{
			name: "GetStores успешный вызов",
			input: args{
				ctx: context.Background(),
				filter: &domain.StoreFilter{
					Limit: 2,
				},
			},
			expectedResult: []*domain.Store{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
				{
					ID:          "00000000-0000-0000-0000-000000000002",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
			},
			expectedError: nil,
		},
		{
			name: "GetStores ошбика выполнения",
			input: args{
				ctx: context.Background(),
				filter: &domain.StoreFilter{
					Limit: 2,
				},
			},
			expectedResult: nil,
			expectedError:  errors.New("custom error"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.EXPECT().
				GetStores(tt.input.ctx, tt.input.filter).
				Return(tt.expectedResult, tt.expectedError)

			store, err := uc.GetStores(tt.input.ctx, tt.input.filter)

			require.Equal(t, tt.expectedError, err)
			require.ElementsMatch(t, tt.expectedResult, store)
		})
	}
}

func TestStoreUsecase_CreateStore(t *testing.T) {
	type args struct {
		ctx         context.Context
		name        string
		description string
		cityID      string
		address     string
		cardImg     string
		openAt      string
		closedAt    string
		rating      float64
	}

	type testCase struct {
		name          string
		input         args
		expectedError error
	}

	tests := []testCase{
		{
			name: "CreateStore успешный вызов",
			input: args{
				ctx:         context.Background(),
				name:        "Store",
				description: "Description",
				cityID:      "10000000-0000-0000-0000-000000000001",
				address:     "Address",
				cardImg:     "CardImg",
				openAt:      "OpenAt",
				closedAt:    "ClosedAt",
				rating:      3,
			},
			expectedError: nil,
		},
		{
			name: "CreateStore ошибка выполнения",
			input: args{
				ctx:         context.Background(),
				name:        "Store",
				description: "Description",
				cityID:      "10000000-0000-0000-0000-000000000001",
				address:     "Address",
				cardImg:     "CardImg",
				openAt:      "OpenAt",
				closedAt:    "ClosedAt",
				rating:      3,
			},
			expectedError: errors.New("custom error"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo.EXPECT().
				CreateStore(tt.input.ctx, &domain.Store{Name: tt.input.name, Description: tt.input.description,
					CityID: tt.input.cityID, Address: tt.input.address, CardImg: tt.input.cardImg, Rating: tt.input.rating,
					OpenAt: tt.input.closedAt, ClosedAt: tt.input.openAt}).
				Return(tt.expectedError)

			err := uc.CreateStore(tt.input.ctx, tt.input.name, tt.input.description, tt.input.cityID,
				tt.input.address, tt.input.cardImg, tt.input.closedAt, tt.input.openAt, tt.input.rating)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
