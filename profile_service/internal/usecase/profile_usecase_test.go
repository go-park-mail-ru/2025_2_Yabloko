package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"apple_backend/profile_service/internal/usecase/mock"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func stringPtr(s string) *string { return &s }

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
		expectRepoCall bool
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
			mockError:      nil,
			expectRepoCall: true,
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
			expectRepoCall: true,
			expectedResult: nil,
			expectedError:  domain.ErrProfileNotFound,
		},
		{
			name:           "Невалидный UUID",
			id:             "invalid-uuid",
			expectRepoCall: false,
			expectedResult: nil,
			expectedError:  domain.ErrInvalidProfileData,
		},
		{
			name:           "Пустой ID",
			id:             "",
			expectRepoCall: false,
			expectedResult: nil,
			expectedError:  domain.ErrInvalidProfileData,
		},
		{
			name:           "Ошибка репозитория",
			id:             "550e8400-e29b-41d4-a716-446655440000",
			mockReturn:     nil,
			mockError:      errors.New("db err"),
			expectRepoCall: true,
			expectedResult: nil,
			expectedError:  errors.New("db err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectRepoCall {
				mockRepo.EXPECT().
					GetProfile(gomock.Any(), tt.id).
					Return(tt.mockReturn, tt.mockError)
			}

			result, err := uc.GetProfile(context.Background(), tt.id)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
			}
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
		name        string
		email       string
		password    string
		setupMock   func(t *testing.T, email, pass string)
		expectError bool
		errorType   error
	}{
		{
			name:     "Успешное создание профиля",
			email:    "newuser@example.com",
			password: "StrongPass123",
			setupMock: func(t *testing.T, email, pass string) {
				mockRepo.EXPECT().
					CreateProfile(gomock.Any(), gomock.AssignableToTypeOf(&domain.Profile{})).
					DoAndReturn(func(ctx context.Context, p *domain.Profile) error {
						require.Equal(t, strings.TrimSpace(email), p.Email)
						require.NotEmpty(t, p.ID)
						_, err := uuid.Parse(p.ID)
						require.NoError(t, err)
						require.NoError(t, bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(pass)))
						return nil
					})
			},
			expectError: false,
		},
		{
			name:     "Email с пробелами, триммим",
			email:    "  test@example.com  ",
			password: "StrongPass123",
			setupMock: func(t *testing.T, email, pass string) {
				mockRepo.EXPECT().
					CreateProfile(gomock.Any(), gomock.AssignableToTypeOf(&domain.Profile{})).
					DoAndReturn(func(ctx context.Context, p *domain.Profile) error {
						require.Equal(t, "test@example.com", p.Email)
						require.NoError(t, bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(pass)))
						return nil
					})
			},
			expectError: false,
		},
		{
			name:        "Пустой email",
			email:       "",
			password:    "StrongPass123",
			setupMock:   func(t *testing.T, email, pass string) {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name:        "Пустой пароль",
			email:       "test@example.com",
			password:    "",
			setupMock:   func(t *testing.T, email, pass string) {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name:        "Невалидный email (без @)",
			email:       "invalid-email",
			password:    "StrongPass123",
			setupMock:   func(t *testing.T, email, pass string) {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name:        "Пароль слишком короткий",
			email:       "test@example.com",
			password:    "1234567",
			setupMock:   func(t *testing.T, email, pass string) {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name:        "Пароль слишком длинный",
			email:       "test@example.com",
			password:    strings.Repeat("p", 73),
			setupMock:   func(t *testing.T, email, pass string) {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name:     "Ошибка репозитория",
			email:    "test@example.com",
			password: "StrongPass123",
			setupMock: func(t *testing.T, email, pass string) {
				mockRepo.EXPECT().
					CreateProfile(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectError: true,
			errorType:   errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(t, tt.email, tt.password)

			id, err := uc.CreateProfile(context.Background(), tt.email, tt.password)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					require.Equal(t, tt.errorType, err)
				}
				require.Empty(t, id)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, id)
			}
		})
	}
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewProfileUsecase(mockRepo)

	newExisting := func() *domain.Profile {
		return &domain.Profile{
			ID:      "550e8400-e29b-41d4-a716-446655440000",
			Email:   "test@example.com",
			Name:    stringPtr("Old Name"),
			Phone:   stringPtr("+79991234567"),
			CityID:  stringPtr("city-123"),
			Address: stringPtr("Old Address"),
		}
	}

	tests := []struct {
		name          string
		updateProfile *domain.Profile
		setupMock     func()
		expectError   bool
		errorType     error
	}{
		{
			name: "Успешное обновление всех полей",
			updateProfile: &domain.Profile{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    stringPtr("New Name"),
				Phone:   stringPtr("+79997654321"),
				CityID:  stringPtr("city-456"),
				Address: stringPtr("New Address"),
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Частичное обновление: только телефон",
			updateProfile: &domain.Profile{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Phone: stringPtr("+79997654321"),
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
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
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Профиль не найден",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440999",
				Name: stringPtr("Name"),
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(nil, domain.ErrProfileNotFound)
			},
			expectError: true,
			errorType:   domain.ErrProfileNotFound,
		},
		{
			name: "Невалидный UUID",
			updateProfile: &domain.Profile{
				ID: "invalid-uuid",
			},
			setupMock:   func() {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Валидация: слишком длинное имя",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr(strings.Repeat("a", 101)),
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
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
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
			},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Валидация: слишком длинный адрес",
			updateProfile: &domain.Profile{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Address: stringPtr(strings.Repeat("x", 201)),
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
			},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
		{
			name: "Ошибка репозитория при UpdateProfile",
			updateProfile: &domain.Profile{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr("New Name"),
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))
			},
			expectError: true,
			errorType:   errors.New("update failed"),
		},
		{
			name: "Пустой профиль (только ID)",
			updateProfile: &domain.Profile{
				ID: "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(newExisting(), nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
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
		errorType   error
	}{
		{
			name: "Успешное удаление",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			setupMock: func() {
				mockRepo.EXPECT().DeleteProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(nil)
			},
			expectError: false,
		},
		{
			name: "Ошибка репозитория",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			setupMock: func() {
				mockRepo.EXPECT().DeleteProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(errors.New("delete error"))
			},
			expectError: true,
			errorType:   errors.New("delete error"),
		},
		{
			name:        "Невалидный UUID",
			id:          "invalid-uuid",
			setupMock:   func() {},
			expectError: true,
			errorType:   domain.ErrInvalidProfileData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := uc.DeleteProfile(context.Background(), tt.id)
			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.errorType, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
