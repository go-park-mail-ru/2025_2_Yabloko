package usecase

import (
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/usecase/mock"
	"context"
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
		repoOutput     []*domain.Store
		expectedResult *domain.StoreAgg
		expectedError  error
	}

	tests := []testCase{
		{
			name: "GetStore успешный вызов",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			repoOutput: []*domain.Store{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagID:       "10000000-0000-0000-0000-000000000001",
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
			},
			expectedResult: &domain.StoreAgg{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "Store",
				Description: "Description",
				CityID:      "10000000-0000-0000-0000-000000000001",
				Address:     "Address",
				CardImg:     "CardImg",
				Rating:      3,
				TagsID:      []string{"10000000-0000-0000-0000-000000000001"},
				OpenAt:      "OpenAt",
				ClosedAt:    "ClosedAt",
			},
			expectedError: nil,
		},
		{
			name: "GetStore успешный вызов несколько категорий",
			input: args{
				ctx: context.Background(),
				id:  "00000000-0000-0000-0000-000000000001",
			},
			repoOutput: []*domain.Store{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagID:       "10000000-0000-0000-0000-000000000001",
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagID:       "10000000-0000-0000-0000-000000000002",
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
			},
			expectedResult: &domain.StoreAgg{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "Store",
				Description: "Description",
				CityID:      "10000000-0000-0000-0000-000000000001",
				Address:     "Address",
				CardImg:     "CardImg",
				Rating:      3,
				TagsID:      []string{"10000000-0000-0000-0000-000000000001", "10000000-0000-0000-0000-000000000002"},
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
			expectedError:  domain.ErrInternalServer,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo.EXPECT().
				GetStore(tt.input.ctx, tt.input.id).
				Return(tt.repoOutput, tt.expectedError)

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
		mockSetup      func(mock *mock.MockStoreRepository, out []*domain.Store, err error)
		repoOutput     []*domain.Store
		expectedResult []*domain.StoreAgg
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
			mockSetup: func(mock *mock.MockStoreRepository, out []*domain.Store, err error) {
				mock.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 2}).
					Return(out, err)
			},
			repoOutput: []*domain.Store{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagID:       "00000000-0000-0000-0000-000000000001",
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
					TagID:       "00000000-0000-0000-0000-000000000001",
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
			},
			expectedResult: []*domain.StoreAgg{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagsID:      []string{"00000000-0000-0000-0000-000000000001"},
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
					TagsID:      []string{"00000000-0000-0000-0000-000000000001"},
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
			},
			expectedError: nil,
		},
		{
			name: "GetStores успешный вызов несколько тегов",
			input: args{
				ctx: context.Background(),
				filter: &domain.StoreFilter{
					Limit: 2,
				},
			},
			mockSetup: func(mock *mock.MockStoreRepository, out []*domain.Store, err error) {
				mock.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 2}).
					Return(out, err)
			},
			repoOutput: []*domain.Store{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagID:       "00000000-0000-0000-0000-000000000001",
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagID:       "00000000-0000-0000-0000-000000000002",
					OpenAt:      "OpenAt",
					ClosedAt:    "ClosedAt",
				},
			},
			expectedResult: []*domain.StoreAgg{
				{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "Store",
					Description: "Description",
					CityID:      "10000000-0000-0000-0000-000000000001",
					Address:     "Address",
					CardImg:     "CardImg",
					Rating:      3,
					TagsID: []string{
						"00000000-0000-0000-0000-000000000001",
						"00000000-0000-0000-0000-000000000002",
					},
					OpenAt:   "OpenAt",
					ClosedAt: "ClosedAt",
				},
			},
			expectedError: nil,
		},
		{
			name: "некорректный параметр сортировки",
			input: args{
				ctx: context.Background(),
				filter: &domain.StoreFilter{
					Sorted: "id",
					Limit:  2,
				},
			},
			mockSetup:      func(mock *mock.MockStoreRepository, out []*domain.Store, err error) {},
			repoOutput:     nil,
			expectedResult: nil,
			expectedError:  domain.ErrRequestParams,
		},
		{
			name: "Limit < 0",
			input: args{
				ctx: context.Background(),
				filter: &domain.StoreFilter{
					Limit: -10,
				},
			},
			mockSetup:      func(mock *mock.MockStoreRepository, out []*domain.Store, err error) {},
			repoOutput:     nil,
			expectedResult: nil,
			expectedError:  domain.ErrRequestParams,
		},
		{
			name: "GetStores ошбика выполнения",
			input: args{
				ctx: context.Background(),
				filter: &domain.StoreFilter{
					Limit: 2,
				},
			},
			mockSetup: func(mock *mock.MockStoreRepository, out []*domain.Store, err error) {
				mock.EXPECT().
					GetStores(context.Background(), &domain.StoreFilter{Limit: 2}).
					Return(out, err)
			},
			repoOutput:     nil,
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockStoreRepository(ctrl)
			tt.mockSetup(mockRepo, tt.repoOutput, tt.expectedError)
			uc := NewStoreUsecase(mockRepo)

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
			expectedError: domain.ErrInternalServer,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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

func TestStoreUsecase_GetCities(t *testing.T) {
	type testCase struct {
		name           string
		expectedResult []*domain.City
		expectedError  error
	}
	ctx := context.Background()
	tests := []testCase{
		{
			name: "успешный вызов",
			expectedResult: []*domain.City{
				{
					ID:   "00000000-0000-0000-0000-000000000001",
					Name: "city1",
				},
				{
					ID:   "00000000-0000-0000-0000-000000000002",
					Name: "city2",
				},
			},
			expectedError: nil,
		},
		{
			name:           "ошбика выполнения",
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo.EXPECT().
				GetCities(ctx).
				Return(tt.expectedResult, tt.expectedError)

			store, err := uc.GetCities(ctx)

			require.Equal(t, tt.expectedError, err)
			require.ElementsMatch(t, tt.expectedResult, store)
		})
	}
}

func TestStoreUsecase_GetTags(t *testing.T) {
	type testCase struct {
		name           string
		expectedResult []*domain.StoreTag
		expectedError  error
	}
	ctx := context.Background()
	tests := []testCase{
		{
			name: "успешный вызов",
			expectedResult: []*domain.StoreTag{
				{
					ID:   "00000000-0000-0000-0000-000000000001",
					Name: "tag1",
				},
				{
					ID:   "00000000-0000-0000-0000-000000000002",
					Name: "tag2",
				},
			},
			expectedError: nil,
		},
		{
			name:           "ошбика выполнения",
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockStoreRepository(ctrl)

	uc := NewStoreUsecase(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo.EXPECT().
				GetTags(ctx).
				Return(tt.expectedResult, tt.expectedError)

			store, err := uc.GetTags(ctx)

			require.Equal(t, tt.expectedError, err)
			require.ElementsMatch(t, tt.expectedResult, store)
		})
	}
}

func TestStoreUsecase_GetStoreReview(t *testing.T) {
	type testCase struct {
		name           string
		inputID        string
		expectedResult []*domain.StoreReview
		expectedError  error
	}

	tests := []testCase{
		{
			name:    "успешный вызов",
			inputID: "00000000-0000-0000-0000-000000000002",
			expectedResult: []*domain.StoreReview{
				{
					UserName:  "Пользователь1",
					Rating:    5,
					Comment:   "хороший товар",
					CreatedAt: "2025-01-01",
				},
				{
					UserName:  "Пользователь2",
					Rating:    5,
					Comment:   "хороший товар",
					CreatedAt: "2025-01-02",
				},
			},
			expectedError: nil,
		},
		{
			name:           "ошбика выполнения",
			inputID:        "00000000-0000-0000-0000-000000000002",
			expectedResult: nil,
			expectedError:  domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mock.NewMockStoreRepository(ctrl)
			uc := NewStoreUsecase(mockRepo)

			mockRepo.EXPECT().
				GetStoreReview(ctx, tt.inputID).
				Return(tt.expectedResult, tt.expectedError)

			store, err := uc.GetStoreReview(ctx, tt.inputID)

			require.Equal(t, tt.expectedError, err)
			require.ElementsMatch(t, tt.expectedResult, store)
		})
	}
}
