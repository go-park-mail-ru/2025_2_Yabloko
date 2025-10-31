package http

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/http/mock"
	"apple_backend/profile_service/internal/delivery/transport"
	"apple_backend/profile_service/internal/domain"
	mock_repository "apple_backend/profile_service/internal/repository/mock"
	"bytes"
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
		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "id2")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Некорректный UUID → 400", func(t *testing.T) {
		mockUC.EXPECT().GetProfile(gomock.Any(), "bad").Return(nil, domain.ErrInvalidProfileData)
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/bad", nil)
		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "bad")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		mockUC.EXPECT().GetProfile(gomock.Any(), "id3").Return(nil, errors.New("internal error"))
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id3", nil)
		w := httptest.NewRecorder()
		handler.GetProfile(w, req, "id3")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)
	handler := NewProfileHandler(mockUC, logger.NewNilLogger(), "/api/v0")

	t.Run("Успешное создание профиля", func(t *testing.T) {
		data := transport.CreateProfileRequest{Email: "a@b.com", Password: "pass123"}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().CreateProfile(gomock.Any(), "a@b.com", "pass123").Return("id1", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.CreateProfile(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var resp transport.CreateProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		require.Equal(t, "id1", resp.ID)
	})

	t.Run("Неверный JSON → 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", bytes.NewReader([]byte("bad")))
		w := httptest.NewRecorder()
		handler.CreateProfile(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Профиль уже существует → 409", func(t *testing.T) {
		data := transport.CreateProfileRequest{Email: "exist@b.com", Password: "pass"}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().CreateProfile(gomock.Any(), "exist@b.com", "pass").Return("", domain.ErrProfileExist)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.CreateProfile(w, req)
		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Ошибка usecase → 400", func(t *testing.T) {
		data := transport.CreateProfileRequest{Email: "bad@b.com", Password: "bad"}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().CreateProfile(gomock.Any(), "bad@b.com", "bad").Return("", domain.ErrInvalidProfileData)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.CreateProfile(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		data := transport.CreateProfileRequest{Email: "x@y.com", Password: "pass"}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().CreateProfile(gomock.Any(), "x@y.com", "pass").Return("", errors.New("err"))

		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.CreateProfile(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Неподдерживаемый метод → 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles", nil)
		w := httptest.NewRecorder()
		handler.CreateProfile(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
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
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id1")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Профиль не найден → 404", func(t *testing.T) {
		data := transport.UpdateProfileRequest{Name: stringPtr("A")}
		body, _ := json.Marshal(data)

		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(domain.ErrProfileNotFound)

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id2", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id2")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Неверный JSON → 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id3", bytes.NewReader([]byte("bad")))
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id3")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Некорректные данные → 400", func(t *testing.T) {
		data := transport.UpdateProfileRequest{Phone: stringPtr("bad")}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(domain.ErrInvalidProfileData)

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id4", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.UpdateProfile(w, req, "id4")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		data := transport.UpdateProfileRequest{Name: stringPtr("X")}
		body, _ := json.Marshal(data)
		mockUC.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(errors.New("err"))

		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id5", bytes.NewReader(body))
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
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "id1")
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Профиль не найден → 404", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "id2").Return(domain.ErrProfileNotFound)
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id2", nil)
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "id2")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Некорректный UUID → 400", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "bad").Return(domain.ErrInvalidProfileData)
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/bad", nil)
		w := httptest.NewRecorder()
		handler.DeleteProfile(w, req, "bad")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неизвестная ошибка сервера → 500", func(t *testing.T) {
		mockUC.EXPECT().DeleteProfile(gomock.Any(), "id3").Return(errors.New("err"))
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id3", nil)
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
		handler.handleProfileRoutes(w, httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id1", nil))
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		updateBody, _ := json.Marshal(transport.UpdateProfileRequest{Name: stringPtr("X")})
		handler.handleProfileRoutes(w, httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id1", bytes.NewReader(updateBody)))
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		handler.handleProfileRoutes(w, httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id1", nil))
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Пустой ID → 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.handleProfileRoutes(w, httptest.NewRequest(http.MethodGet, "/api/v0/profiles/", nil))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Неподдерживаемый метод → 405", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.handleProfileRoutes(w, httptest.NewRequest(http.MethodPatch, "/api/v0/profiles/id1", nil))
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestExtractIDFromRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUC := mock.NewMockProfileUsecaseInterface(ctrl)

	h := NewProfileHandler(mockUC, logger.NewNilLogger(), "/api/v0")

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v0/profiles/id1", "id1"},
		{"/api/v0/profiles/id1/", "id1"},
		{"/api/v0/profiles/", ""},
		{"/api/v0/profiles////", ""},
		{"/api/v0/profiles/123/extra/path", "123"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		require.Equal(t, tt.expected, h.extractIDFromRequest(req))
	}
}

func TestNewProfileRouter(t *testing.T) {
	mux := http.NewServeMux()
	NewProfileRouter(mux, nil, "/api/v0", logger.NewNilLogger(), "/upload", "http://base.url")

	req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusNotFound, w.Code)

	reqAvatar := httptest.NewRequest(http.MethodPost, "/api/v0/profiles/id/avatar", nil)
	wAvatar := httptest.NewRecorder()
	mux.ServeHTTP(wAvatar, reqAvatar)
	require.NotEqual(t, http.StatusNotFound, wAvatar.Code)
}

func TestNewProfileRouter_Routes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPgx := mock_repository.NewMockPgxIface(ctrl)
	log := logger.NewNilLogger()
	mux := http.NewServeMux()
	uploadPath := "/tmp/uploads"
	baseURL := "http://localhost:8080"
	apiPrefix := "/api/v0"

	NewProfileRouter(mux, mockPgx, apiPrefix, log, uploadPath, baseURL)

	t.Run("POST /profiles", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles", bytes.NewReader([]byte(`{"email":"a@b.com","password":"pass"}`)))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("GET /profiles/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/id1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /profiles/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v0/profiles/id1", bytes.NewReader([]byte(`{"name":"X"}`)))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("DELETE /profiles/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v0/profiles/id1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("POST /profiles/{id}/avatar", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v0/profiles/id1/avatar", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("PATCH /profiles/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v0/profiles/id1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET /profiles/ (пустой ID)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v0/profiles/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
