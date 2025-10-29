package http

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/http/mock"
	"apple_backend/profile_service/internal/delivery/transport"
	"apple_backend/profile_service/internal/domain"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestProfileHandler_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Успешное получение профиля", func(t *testing.T) {
		expectedProfile := &domain.Profile{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Email: "test@example.com",
			Name:  stringPtr("John"),
		}

		mockUC.EXPECT().
			GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(expectedProfile, nil)

		req := httptest.NewRequest("GET", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusOK, w.Code)

		var response transport.ProfileResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.ID)
		require.Equal(t, "test@example.com", response.Email)
	})

	t.Run("Профиль не найден", func(t *testing.T) {
		mockUC.EXPECT().
			GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(nil, domain.ErrProfileNotFound)

		req := httptest.NewRequest("GET", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Ошибка сервера при получении профиля", func(t *testing.T) {
		mockUC.EXPECT().
			GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(nil, errors.New("internal error"))

		req := httptest.NewRequest("GET", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Некорректный UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v0/profiles/invalid-uuid", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req, "invalid-uuid")

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_GetProfileByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Успешное получение профиля по email", func(t *testing.T) {
		expectedProfile := &domain.Profile{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Email: "test@example.com",
			Name:  stringPtr("John"),
		}

		mockUC.EXPECT().
			GetProfileByEmail(gomock.Any(), "test@example.com").
			Return(expectedProfile, nil)

		req := httptest.NewRequest("GET", "/api/v0/profiles/email/test@example.com", nil)
		w := httptest.NewRecorder()

		handler.GetProfileByEmail(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response transport.ProfileResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "test@example.com", response.Email)
	})

	t.Run("Пустой email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v0/profiles/email/", nil)
		w := httptest.NewRecorder()

		handler.GetProfileByEmail(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Профиль не найден по email", func(t *testing.T) {
		mockUC.EXPECT().
			GetProfileByEmail(gomock.Any(), "notfound@example.com").
			Return(nil, domain.ErrProfileNotFound)

		req := httptest.NewRequest("GET", "/api/v0/profiles/email/notfound@example.com", nil)
		w := httptest.NewRecorder()

		handler.GetProfileByEmail(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Неподдерживаемый метод", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v0/profiles/email/test@example.com", nil)
		w := httptest.NewRecorder()

		handler.GetProfileByEmail(w, req)

		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Ошибка сервера при получении по email", func(t *testing.T) {
		mockUC.EXPECT().
			GetProfileByEmail(gomock.Any(), "error@example.com").
			Return(nil, errors.New("internal error"))

		req := httptest.NewRequest("GET", "/api/v0/profiles/email/error@example.com", nil)
		w := httptest.NewRecorder()

		handler.GetProfileByEmail(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Успешное обновление профиля", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{
			Name:    stringPtr("Updated Name"),
			Phone:   stringPtr("+123456789"),
			CityID:  stringPtr("city-123"),
			Address: stringPtr("New Address"),
		}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, profile *domain.Profile) error {
				require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", profile.ID)
				require.Equal(t, "Updated Name", *profile.Name)
				require.Equal(t, "+123456789", *profile.Phone)
				return nil
			})

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "Профиль успешно обновлен", response["message"])
	})

	t.Run("Частичное обновление профиля", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{
			Name: stringPtr("Only Name Updated"),
		}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, profile *domain.Profile) error {
				require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", profile.ID)
				require.Equal(t, "Only Name Updated", *profile.Name)
				require.Nil(t, profile.Phone)
				require.Nil(t, profile.CityID)
				require.Nil(t, profile.Address)
				return nil
			})

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Обновление только телефона", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{
			Phone: stringPtr("+987654321"),
		}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, profile *domain.Profile) error {
				require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", profile.ID)
				require.Nil(t, profile.Name)
				require.Equal(t, "+987654321", *profile.Phone)
				return nil
			})

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Неверный JSON при обновлении", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Профиль не найден при обновлении", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{
			Name: stringPtr("Updated Name"),
		}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			Return(domain.ErrProfileNotFound)

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Ошибка сервера при обновлении", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{
			Name: stringPtr("Updated Name"),
		}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			Return(errors.New("internal error"))

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Пустой запрос на обновление", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, profile *domain.Profile) error {
				require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", profile.ID)
				require.Nil(t, profile.Name)
				require.Nil(t, profile.Phone)
				require.Nil(t, profile.CityID)
				require.Nil(t, profile.Address)
				return nil
			})

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestProfileHandler_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Успешное удаление профиля", func(t *testing.T) {
		mockUC.EXPECT().
			DeleteProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Профиль не найден при удалении", func(t *testing.T) {
		mockUC.EXPECT().
			DeleteProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(domain.ErrProfileNotFound)

		req := httptest.NewRequest("DELETE", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Ошибка сервера при удалении", func(t *testing.T) {
		mockUC.EXPECT().
			DeleteProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(errors.New("internal error"))

		req := httptest.NewRequest("DELETE", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_HandleProfileRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Неподдерживаемый HTTP метод", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.handleProfileRoutes(w, req)

		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET запрос через handleProfileRoutes", func(t *testing.T) {
		expectedProfile := &domain.Profile{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Email: "test@example.com",
		}

		mockUC.EXPECT().
			GetProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(expectedProfile, nil)

		req := httptest.NewRequest("GET", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.handleProfileRoutes(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PUT запрос через handleProfileRoutes", func(t *testing.T) {
		updateData := transport.UpdateProfileRequest{
			Name: stringPtr("Updated Name"),
		}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			Return(nil)

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.handleProfileRoutes(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DELETE запрос через handleProfileRoutes", func(t *testing.T) {
		mockUC.EXPECT().
			DeleteProfile(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
			Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.handleProfileRoutes(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Неверный UUID в пути", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v0/profiles/invalid-uuid", nil)
		w := httptest.NewRecorder()

		handler.handleProfileRoutes(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Успешное создание профиля", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "newuser@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(createData)

		mockUC.EXPECT().
			CreateProfile(gomock.Any(), "newuser@example.com", "password123").
			Return("new-profile-id", nil)

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response transport.CreateProfileResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "new-profile-id", response.ID)
	})

	t.Run("Неверный JSON при создании", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Профиль уже существует", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(createData)

		mockUC.EXPECT().
			CreateProfile(gomock.Any(), "existing@example.com", "password123").
			Return("", domain.ErrProfileExist)

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Пустой email", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "",
			Password: "password123",
		}
		body, _ := json.Marshal(createData)

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Пустой password", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "test@example.com",
			Password: "",
		}
		body, _ := json.Marshal(createData)

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Ошибка сервера при создании", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(createData)

		mockUC.EXPECT().
			CreateProfile(gomock.Any(), "test@example.com", "password123").
			Return("", errors.New("internal error"))

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Неподдерживаемый метод для CreateProfile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v0/profiles", nil)
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Невалидный email формат", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "invalid-email",
			Password: "password123",
		}
		body, _ := json.Marshal(createData)

		mockUC.EXPECT().
			CreateProfile(gomock.Any(), "invalid-email", "password123").
			Return("", domain.ErrInvalidProfileData)

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Короткий пароль", func(t *testing.T) {
		createData := transport.CreateProfileRequest{
			Email:    "test@example.com",
			Password: "123",
		}
		body, _ := json.Marshal(createData)

		mockUC.EXPECT().
			CreateProfile(gomock.Any(), "test@example.com", "123").
			Return("", domain.ErrInvalidProfileData)

		req := httptest.NewRequest("POST", "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestExtractIDFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		prefix   string
		expected string
	}{
		{
			name:     "Успешное извлечение ID",
			path:     "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000",
			prefix:   "/api/v0/profiles/",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "Извлечение email",
			path:     "/api/v0/profiles/email/test@example.com",
			prefix:   "/api/v0/profiles/email/",
			expected: "test@example.com",
		},
		{
			name:     "Путь с завершающим слешем",
			path:     "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000/",
			prefix:   "/api/v0/profiles/",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "Пустой путь",
			path:     "/api/v0/profiles/",
			prefix:   "/api/v0/profiles/",
			expected: "",
		},
		{
			name:     "Только префикс",
			path:     "/api/v0/profiles/",
			prefix:   "/api/v0/profiles/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIDFromPath(tt.path, tt.prefix)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNewProfileRouter(t *testing.T) {
	mux := http.NewServeMux()
	appLog := logger.NewNilLogger()

	NewProfileRouter(mux, nil, "/api/v0", appLog)

	require.NotNil(t, mux)
}

func stringPtr(s string) *string {
	return &s
}
