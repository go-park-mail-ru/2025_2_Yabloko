//go:generate mockgen -source=profile_handler.go -destination=mock/profile_usecase_mock.go -package=mock
package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/transport"
	"apple_backend/profile_service/internal/domain"
	"apple_backend/profile_service/internal/repository"
	"apple_backend/profile_service/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type ProfileUsecaseInterface interface {
	GetProfile(ctx context.Context, id string) (*domain.Profile, error)
	GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error)
	CreateProfile(ctx context.Context, email, passwordHash string) (string, error)
	UpdateProfile(ctx context.Context, profile *domain.Profile) error
	DeleteProfile(ctx context.Context, id string) error
}

type ProfileHandler struct {
	uc ProfileUsecaseInterface
	rs *http_response.ResponseSender
}

func NewProfileHandler(uc ProfileUsecaseInterface, log *logger.Logger) *ProfileHandler {
	return &ProfileHandler{
		uc: uc,
		rs: http_response.NewResponseSender(log),
	}
}

func NewProfileRouter(mux *http.ServeMux, db repository.PgxIface, apiPrefix string, appLog *logger.Logger) {
	profileRepo := repository.NewProfileRepoPostgres(db, appLog)
	profileUC := usecase.NewProfileUsecase(profileRepo)
	profileHandler := NewProfileHandler(profileUC, appLog)

	// POST /apiPrefix/profiles Create (без id)
	mux.HandleFunc(apiPrefix+"/profiles", profileHandler.CreateProfile)

	// GET/PUT/DELETE /apiPrefix/profiles/{id}
	mux.HandleFunc(apiPrefix+"/profiles/", profileHandler.handleProfileRoutes)

	// GET /apiPrefix/profiles/email/{email}
	mux.HandleFunc(apiPrefix+"/profiles/email/", profileHandler.GetProfileByEmail)
}

func extractIDFromPath(path, prefix string) string {
	id := strings.TrimPrefix(path, prefix)
	return strings.TrimSuffix(id, "/")
}

func (h *ProfileHandler) handleProfileRoutes(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v0/profiles/")

	switch r.Method {
	case http.MethodGet:
		h.GetProfile(w, r, id)
	case http.MethodPut:
		h.UpdateProfile(w, r, id)
	case http.MethodDelete:
		h.DeleteProfile(w, r, id)
	default:
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "ProfileRoutes", domain.ErrHTTPMethod, nil)
	}
}

// GetProfile godoc
// @Summary Получить профиль по ID
// @Description Возвращает информацию о профиле пользователя по его UUID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "UUID пользователя"
// @Success 200 {object} transport.ProfileResponse
// @Failure 400 {object} http_response.ErrResponse "Некорректный ID"
// @Failure 404 {object} http_response.ErrResponse "Профиль не найден"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /profiles/{id} [get]
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request, id string) {
	profile, err := h.uc.GetProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetProfile", domain.ErrProfileNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetProfile", domain.ErrInternalServer, err)
		return
	}

	responseProfile := transport.ToProfileResponse(profile)
	h.rs.Send(r.Context(), w, http.StatusOK, responseProfile)
}

// GetProfileByEmail godoc
// @Summary Получить профиль по email
// @Description Возвращает информацию о профиле пользователя по его email
// @Tags profiles
// @Accept json
// @Produce json
// @Param email path string true "Email пользователя"
// @Success 200 {object} transport.ProfileResponse
// @Failure 400 {object} http_response.ErrResponse "Некорректный email"
// @Failure 404 {object} http_response.ErrResponse "Профиль не найден"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /profiles/email/{email} [get]
func (h *ProfileHandler) GetProfileByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetProfileByEmail", domain.ErrHTTPMethod, nil)
		return
	}

	email := extractIDFromPath(r.URL.Path, "/api/v0/profiles/email/")
	if email == "" {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetProfileByEmail", domain.ErrRequestParams, nil)
		return
	}

	profile, err := h.uc.GetProfileByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetProfileByEmail", domain.ErrProfileNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetProfileByEmail", domain.ErrInternalServer, err)
		return
	}

	responseProfile := transport.ToProfileResponse(profile)
	h.rs.Send(r.Context(), w, http.StatusOK, responseProfile)
}

// CreateProfile godoc
// @Summary Создать профиль
// @Description Создает новый профиль пользователя
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body transport.CreateProfileRequest true "Данные для создания профиля"
// @Success 201 {object} transport.CreateProfileResponse
// @Failure 400 {object} http_response.ErrResponse "Ошибка входных данных"
// @Failure 409 {object} http_response.ErrResponse "Профиль уже существует"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /profiles [post]
func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "CreateProfile", domain.ErrHTTPMethod, nil)
		return
	}

	var req transport.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "CreateProfile", domain.ErrRequestParams, err)
		return
	}

	profileID, err := h.uc.CreateProfile(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrProfileExist) {
			h.rs.Error(r.Context(), w, http.StatusConflict, "CreateProfile", domain.ErrProfileExist, err)
			return
		}
		if errors.Is(err, domain.ErrInvalidProfileData) {
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "CreateProfile", domain.ErrInvalidProfileData, err)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "CreateProfile", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusCreated, &transport.CreateProfileResponse{ID: profileID})
}

// UpdateProfile godoc
// @Summary Обновить профиль
// @Description Обновляет информацию о профиле пользователя
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "UUID пользователя"
// @Param request body transport.UpdateProfileRequest true "Данные для обновления профиля"
// @Success 200 {object} map[string]string "Профиль успешно обновлен"
// @Failure 400 {object} http_response.ErrResponse "Ошибка входных данных"
// @Failure 404 {object} http_response.ErrResponse "Профиль не найден"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /profiles/{id} [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request, id string) {
	var req transport.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateProfile", domain.ErrRequestParams, err)
		return
	}

	profile := &domain.Profile{
		ID:      id,
		Name:    req.Name,
		Phone:   req.Phone,
		CityID:  req.CityID,
		Address: req.Address,
	}

	err := h.uc.UpdateProfile(r.Context(), profile)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "UpdateProfile", domain.ErrProfileNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "UpdateProfile", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusOK, map[string]string{"message": "Профиль успешно обновлен"})
}

// DeleteProfile godoc
// @Summary Удалить профиль
// @Description Удаляет профиль пользователя по UUID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "UUID пользователя"
// @Success 204 "Профиль успешно удален"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Профиль не найден"
// @Failure 405 {object} map[string]string "Неверный HTTP-метод"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /profiles/{id} [delete]
func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request, id string) {
	err := h.uc.DeleteProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			h.rs.Error(r.Context(), w, http.StatusNotFound, "DeleteProfile", domain.ErrProfileNotFound, nil)
			return
		}
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "DeleteProfile", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusNoContent, nil)
}
