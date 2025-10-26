package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"apple_backend/profile_service/internal/usecase/mock"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestProfileUsecase_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	tests := []struct {
		name           string
		id             string
		mockReturn     *domain.Profile
		mockError      error
		expectedResult *domain.Profile
		expectedError  error
	}{
		{
			name: "Успешное получение профиля",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			mockReturn: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Email: "test@example.com",
				Name:  stringPtr("John"),
			},
			mockError: nil,
			expectedResult: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Email: "test@example.com",
				Name:  stringPtr("John"),
			},
			expectedError: nil,
		},
		{
			name:           "Профиль не найден",
			id:             "550e8400-e29b-41d4-a716-446655440000",
			mockReturn:     nil,
			mockError:      domain.ErrProfileNotFound,
			expectedResult: nil,
			expectedError:  domain.ErrProfileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetProfile(gomock.Any(), tt.id).
				Return(tt.mockReturn, tt.mockError)

			result, err := uc.GetProfile(context.Background(), tt.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	profile := &domain.Profile{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: stringPtr("Updated Name"),
	}

	t.Run("Успешное обновление", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), profile).
			Return(nil)

		err := uc.UpdateProfile(context.Background(), profile)
		require.NoError(t, err)
	})

	t.Run("Ошибка при обновлении", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), profile).
			Return(errors.New("update error"))

		err := uc.UpdateProfile(context.Background(), profile)
		require.Error(t, err)
	})
}

func TestProfileUsecase_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	t.Run("Успешное удаление", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), "test-id").
			Return(nil)

		err := uc.DeleteProfile(context.Background(), "test-id")
		require.NoError(t, err)
	})

	t.Run("Ошибка при удалении", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), "test-id").
			Return(errors.New("delete error"))

		err := uc.DeleteProfile(context.Background(), "test-id")
		require.Error(t, err)
	})
}

func TestProfileUsecase_GetProfileByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	tests := []struct {
		name           string
		email          string
		mockReturn     *domain.Profile
		mockError      error
		expectedResult *domain.Profile
		expectedError  error
	}{
		{
			name:  "Успешное получение профиля по email",
			email: "test@example.com",
			mockReturn: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Email: "test@example.com",
				Name:  stringPtr("John"),
			},
			mockError: nil,
			expectedResult: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Email: "test@example.com",
				Name:  stringPtr("John"),
			},
			expectedError: nil,
		},
		{
			name:           "Профиль не найден по email",
			email:          "notfound@example.com",
			mockReturn:     nil,
			mockError:      domain.ErrProfileNotFound,
			expectedResult: nil,
			expectedError:  domain.ErrProfileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetProfileByEmail(gomock.Any(), tt.email).
				Return(tt.mockReturn, tt.mockError)

			result, err := uc.GetProfileByEmail(context.Background(), tt.email)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
