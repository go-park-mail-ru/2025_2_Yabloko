package http

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/mock"
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

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response["id"])
		require.Equal(t, "test@example.com", response["email"])
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

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "test@example.com", response["email"])
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
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	appLog := logger.NewNilLogger()
	handler := NewProfileHandler(mockUC, appLog)

	t.Run("Успешное обновление профиля", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name":    "Updated Name",
			"phone":   "+123456789",
			"city_id": "city-123",
			"address": "New Address",
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
	})

	t.Run("Неверный JSON при обновлении", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Профиль не найден при обновлении", func(t *testing.T) {
		updateData := map[string]interface{}{"name": "Updated Name"}
		body, _ := json.Marshal(updateData)

		mockUC.EXPECT().
			UpdateProfile(gomock.Any(), gomock.Any()).
			Return(domain.ErrProfileNotFound)

		req := httptest.NewRequest("PUT", "/api/v0/profiles/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req, "550e8400-e29b-41d4-a716-446655440000")

		require.Equal(t, http.StatusNotFound, w.Code)
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

	t.Run("Неверный UUID в пути", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v0/profiles/invalid-uuid", nil)
		w := httptest.NewRecorder()

		handler.handleProfileRoutes(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func stringPtr(s string) *string {
	return &s
}
