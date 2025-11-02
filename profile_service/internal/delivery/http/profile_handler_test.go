package http

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/http/mock"
	"apple_backend/profile_service/internal/delivery/middlewares" // добавлено
	"apple_backend/profile_service/internal/delivery/transport"
	"apple_backend/profile_service/internal/domain"
	"bytes"
	"context" // добавлено
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string { return &s }

func TestProfileHandler_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	handler := NewProfileHandler(mockUC, logger.NewNilLogger(), "/api/v0")

	t.Run("Успешное получение профиля", func(t *testing.T) {
		profile := &domain.Profile{ID: "id1", Email: "test@example.com", Name: stringPtr("John")}
		mockUC.EXPECT().GetProfile(gomock.Any(), "id1").Return(profile, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено

		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "id1")

		require.Equal(t, http.StatusOK, w.Code)
		var resp transport.ProfileResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Equal(t, "id1", resp.ID)
	})

	t.Run("Профиль не найден → 404", func(t *testing.T) {
		mockUC.EXPECT().GetProfile(gomock.Any(), "id2").Return(nil, domain.ErrProfileNotFound)
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id2", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id2")) // добавлено
		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "id2")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Некорректный UUID → 400", func(t *testing.T) {
		mockUC.EXPECT().GetProfile(gomock.Any(), "bad").Return(nil, domain.ErrInvalidProfileData)
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/bad", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "bad")) // добавлено
		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "bad")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		mockUC.EXPECT().GetProfile(gomock.Any(), "id3").Return(nil, errors.New("internal error"))
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id3", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id3")) // добавлено
		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "id3")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	handler := NewProfileHandler(mockUC, logger.NewNilLogger(), "/api/v0")

	t.Run("Успешное обновление всех полей", func(t *testing.T) {
		data := transport.UpdateProfileRequest{
			Name:    stringPtr("John"),
			Phone:   stringPtr("+123"),
			CityID:  stringPtr("cid"),
			Address: stringPtr("addr"),
		}
		body, _ := json.Marshal(data)

		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id1", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id1")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Профиль не найден → 404", func(t *testing.T) {
		data := transport.UpdateProfileRequest{Name: stringPtr("A")}
		body, _ := json.Marshal(data)

		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(domain.ErrProfileNotFound)

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id2", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id2")) // добавлено
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id2")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Неверный JSON → 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id3", bytes.NewReader([]byte("bad")))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id3")) // добавлено
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id3")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Некорректные данные → 400", func(t *testing.T) {
		data := transport.UpdateProfileRequest{Phone: stringPtr("bad")}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(domain.ErrInvalidProfileData)

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id4", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id4")) // добавлено
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id4")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		data := transport.UpdateProfileRequest{Name: stringPtr("X")}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(errors.New("err"))

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id5", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id5")) // добавлено
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id5")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	handler := NewProfileHandler(mockUC, logger.NewNilLogger(), "/api/v0")

	t.Run("Успешное удаление → 204", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "id1").Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "id1")
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Профиль не найден → 404", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "id2").Return(domain.ErrProfileNotFound)
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id2", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id2")) // добавлено
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "id2")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Некорректный UUID → 400", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "bad").Return(domain.ErrInvalidProfileData)
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/bad", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "bad")) // добавлено
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "bad")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "id3").Return(errors.New("err"))
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id3", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id3")) // добавлено
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "id3")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_HandleProfileRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	handler := NewProfileHandler(mockUC, logger.NewNilLogger(), "/api/v0")

	t.Run("GET/PUT/DELETE через handleProfileRoutes", func(t *testing.T) {
		mockUC.EXPECT().GetProfile(gomock.Any(), "id1").Return(&domain.Profile{ID: "id1", Email: "e"}, nil)
		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "id1").Return(nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено
		handler.handleProfileRoutes(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		updateBody, _ := json.Marshal(transport.UpdateProfileRequest{Name: stringPtr("X")})
		req = httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id1", bytes.NewReader(updateBody))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено
		handler.handleProfileRoutes(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено
		handler.handleProfileRoutes(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Пустой ID → 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.handleProfileRoutes(w, httptest.NewRequest(http.MethodGet, "/api/v0/profiles/", nil))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неподдерживаемый метод → 405", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v0/profiles/id1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.CtxUserID, "id1")) // добавлено
		handler.handleProfileRoutes(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
