package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"apple_backend/profile_service/internal/usecase/mock"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestProfileUsecase_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	t.Run("Успешное создание профиля", func(t *testing.T) {
		email := "newuser@example.com"
		passwordHash := "hashedpassword123"

		mockRepo.EXPECT().
			GetProfileByEmail(gomock.Any(), email).
			Return(nil, domain.ErrProfileNotFound)

		mockRepo.EXPECT().
			CreateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, profile *domain.Profile) error {
				require.NotEmpty(t, profile.ID)
				require.Equal(t, email, profile.Email)
				require.Equal(t, passwordHash, profile.PasswordHash)
				return nil
			})

		result, err := uc.CreateProfile(context.Background(), email, passwordHash)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, email, result.Email)
		require.Equal(t, passwordHash, result.PasswordHash)
	})

	t.Run("Профиль уже существует", func(t *testing.T) {
		email := "existing@example.com"
		passwordHash := "hashedpassword123"

		mockRepo.EXPECT().
			GetProfileByEmail(gomock.Any(), email).
			Return(&domain.Profile{ID: "some-id", Email: email}, nil)

		result, err := uc.CreateProfile(context.Background(), email, passwordHash)

		require.Error(t, err)
		require.Equal(t, domain.ErrProfileExist, err)
		require.Nil(t, result)
	})

	t.Run("Ошибка БД при проверке email", func(t *testing.T) {
		email := "test@example.com"
		passwordHash := "hashedpassword123"
		dbError := errors.New("database connection error")

		mockRepo.EXPECT().
			GetProfileByEmail(gomock.Any(), email).
			Return(nil, dbError)

		result, err := uc.CreateProfile(context.Background(), email, passwordHash)

		require.Error(t, err)
		require.Equal(t, dbError, err)
		require.Nil(t, result)
	})

	t.Run("Ошибка при создании профиля в БД", func(t *testing.T) {
		email := "newuser@example.com"
		passwordHash := "hashedpassword123"
		dbError := errors.New("insert failed")

		mockRepo.EXPECT().
			GetProfileByEmail(gomock.Any(), email).
			Return(nil, domain.ErrProfileNotFound)

		mockRepo.EXPECT().
			CreateProfile(gomock.Any(), gomock.Any()).
			Return(dbError)

		result, err := uc.CreateProfile(context.Background(), email, passwordHash)

		require.Error(t, err)
		require.Equal(t, dbError, err)
		require.Nil(t, result)
	})
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	existingProfile := &domain.Profile{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Email:     "test@example.com",
		Name:      stringPtr("Old Name"),
		Phone:     stringPtr("+79991234567"),
		Address:   stringPtr("Old Address"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Успешное обновление профиля", func(t *testing.T) {
		updateProfile := &domain.Profile{
			ID:   "550e8400-e29b-41d4-a716-446655440000",
			Name: stringPtr("New Name"),
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, profile *domain.Profile) error {
				require.Equal(t, "New Name", *profile.Name)
				require.Equal(t, "+79991234567", *profile.Phone)
				require.Equal(t, "Old Address", *profile.Address)
				return nil
			})

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.NoError(t, err)
	})

	t.Run("Профиль не найден при обновлении", func(t *testing.T) {
		updateProfile := &domain.Profile{
			ID:   "non-existent-id",
			Name: stringPtr("New Name"),
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(nil, domain.ErrProfileNotFound)

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.Error(t, err)
		require.Equal(t, domain.ErrProfileNotFound, err)
	})

	t.Run("Валидация: слишком длинное имя", func(t *testing.T) {
		longName := strings.Repeat("a", 101)
		updateProfile := &domain.Profile{
			ID:   "550e8400-e29b-41d4-a716-446655440000",
			Name: &longName,
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.Error(t, err)
		require.Equal(t, domain.ErrInvalidProfileData, err)
	})

	t.Run("Валидация: слишком короткий телефон", func(t *testing.T) {
		shortPhone := "123"
		updateProfile := &domain.Profile{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Phone: &shortPhone,
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.Error(t, err)
		require.Equal(t, domain.ErrInvalidProfileData, err)
	})

	t.Run("Валидация: слишком длинный телефон", func(t *testing.T) {
		longPhone := strings.Repeat("1", 21)
		updateProfile := &domain.Profile{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Phone: &longPhone,
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.Error(t, err)
		require.Equal(t, domain.ErrInvalidProfileData, err)
	})

	t.Run("Валидация: слишком длинный адрес", func(t *testing.T) {
		longAddress := strings.Repeat("a", 201)
		updateProfile := &domain.Profile{
			ID:      "550e8400-e29b-41d4-a716-446655440000",
			Address: &longAddress,
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.Error(t, err)
		require.Equal(t, domain.ErrInvalidProfileData, err)
	})

	t.Run("Частичное обновление: только телефон", func(t *testing.T) {
		newPhone := "+79997654321"
		updateProfile := &domain.Profile{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Phone: &newPhone,
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, profile *domain.Profile) error {
				require.Equal(t, "Old Name", *profile.Name)
				require.Equal(t, "+79997654321", *profile.Phone)
				require.Equal(t, "Old Address", *profile.Address)
				return nil
			})

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.NoError(t, err)
	})

	t.Run("Ошибка БД при обновлении", func(t *testing.T) {
		updateProfile := &domain.Profile{
			ID:   "550e8400-e29b-41d4-a716-446655440000",
			Name: stringPtr("New Name"),
		}
		dbError := errors.New("update failed")

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), updateProfile.ID).
			Return(existingProfile, nil)

		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			Return(dbError)

		err := uc.UpdateProfile(context.Background(), updateProfile)
		require.Error(t, err)
		require.Equal(t, dbError, err)
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

func stringPtr(s string) *string {
	return &s
}
