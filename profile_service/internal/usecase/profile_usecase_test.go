package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"apple_backend/profile_service/internal/usecase/mock"
	"context"
	"errors"
	"strings"
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
		{
			name:           "Невалидный UUID",
			id:             "invalid-uuid",
			mockReturn:     nil,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  domain.ErrInvalidProfileData,
		},
		{
			name:           "Пустой ID",
			id:             "",
			mockReturn:     nil,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  domain.ErrInvalidProfileData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError != domain.ErrInvalidProfileData {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), tt.id).
					Return(tt.mockReturn, tt.mockError)
			}

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
		{
			name:           "Ошибка репозитория",
			email:          "error@example.com",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedResult: nil,
			expectedError:  errors.New("database error"),
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

	tests := []struct {
		name         string
		email        string
		passwordHash string
		setupMock    func()
		expectError  bool
		errorType    error
	}{
		{
			name:         "Успешное создание профиля",
			email:        "newuser@example.com",
			passwordHash: "hashedpassword123",
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:         "Пустой email",
			email:        "",
			passwordHash: "hashedpassword123",
			setupMock:    func() {},
			expectError:  true,
			errorType:    domain.ErrInvalidProfileData,
		},
		{
			name:         "Пустой password hash",
			email:        "test@example.com",
			passwordHash: "",
			setupMock:    func() {},
			expectError:  true,
			errorType:    domain.ErrInvalidProfileData,
		},
		{
			name:         "Email с пробелами",
			email:        "  test@example.com  ",
			passwordHash: "hashedpassword123",
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:         "Ошибка БД при создании",
			email:        "test@example.com",
			passwordHash: "hashedpassword123",
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProfile(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectError: true,
			errorType:   errors.New("database error"),
		},
		{
			name:         "Email только из пробелов",
			email:        "   ",
			passwordHash: "hashedpassword123",
			setupMock:    func() {},
			expectError:  true,
			errorType:    domain.ErrInvalidProfileData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			profileID, err := uc.CreateProfile(context.Background(), tt.email, tt.passwordHash)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					require.Equal(t, tt.errorType, err)
				}
				require.Empty(t, profileID)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, profileID)
			}
		})
	}
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	existingProfile := &domain.Profile{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Email:   "test@example.com",
		Name:    stringPtr("Old Name"),
		Phone:   stringPtr("+79991234567"),
		CityID:  stringPtr("city-123"),
		Address: stringPtr("Old Address"),
	}

	tests := []struct {
		name          string
		updateProfile *domain.Profile
		setupMock     func()
		expectError   bool
		errorType     error
	}{
		{
			name: "Успешное обновление профиля",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr("New Name"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Профиль не найден при обновлении",
			updateProfile: &domain.Profile{
				ID:   "non-existent-id",
				Name: stringPtr("New Name"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "non-existent-id").
					Return(nil, domain.ErrProfileNotFound)
			},
			expectError: true,
			errorType:   domain.ErrProfileNotFound,
		},
		{
			name: "Валидация: слишком длинное имя",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr(strings.Repeat("a", 101)),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
			},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Валидация: слишком короткий телефон",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr("123"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
			},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Валидация: слишком длинный телефон",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr(strings.Repeat("1", 21)),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
			},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Валидация: слишком длинный адрес",
			updateProfile: &domain.Profile{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Address: stringPtr(strings.Repeat("a", 201)),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
			},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Частичное обновление: только телефон",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr("+79997654321"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Обновление CityID",
			updateProfile: &domain.Profile{
				ID:     "550e8400-e29b-41d4-a716-446655440000",
				CityID: stringPtr("city-456"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Ошибка БД при обновлении",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr("New Name"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed"))
			},
			expectError: true,
			errorType:   errors.New("update failed"),
		},
		{
			name: "Обновление всех полей",
			updateProfile: &domain.Profile{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    stringPtr("New Name"),
				Phone:   stringPtr("+79997654321"),
				CityID:  stringPtr("city-456"),
				Address: stringPtr("New Address"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Телефон с пробелами",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr("  +79991234567  "),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Минимальная длина телефона",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr("1234567890"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Максимальная длина телефона",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr("+1234567890123456789"),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Максимальная длина имени",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr(strings.Repeat("a", 100)),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Максимальная длина адреса",
			updateProfile: &domain.Profile{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Address: stringPtr(strings.Repeat("a", 200)),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(existingProfile, nil)
				mockRepo.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := uc.UpdateProfile(context.Background(), tt.updateProfile)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.errorType, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProfileUsecase_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	tests := []struct {
		name        string
		id          string
		setupMock   func()
		expectError bool
	}{
		{
			name: "Успешное удаление",
			id:   "test-id",
			setupMock: func() {
				mockRepo.EXPECT().
					DeleteProfile(gomock.Any(), "test-id").
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "Ошибка при удалении",
			id:   "test-id",
			setupMock: func() {
				mockRepo.EXPECT().
					DeleteProfile(gomock.Any(), "test-id").
					Return(errors.New("delete error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := uc.DeleteProfile(context.Background(), tt.id)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
