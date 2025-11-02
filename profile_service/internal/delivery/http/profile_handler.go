//go:generate mockgen -source=profile_handler.go -destination=mock/profile_usecase_mock.go -package=mock
package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/middlewares"
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
	UpdateProfile(ctx context.Context, profile *domain.Profile) error
	DeleteProfile(ctx context.Context, id string) error
}

type ProfileHandler struct {
	uc           ProfileUsecaseInterface
	rs           *http_response.ResponseSender
	profilesPath string
}

func NewProfileHandler(uc ProfileUsecaseInterface, log *logger.Logger, apiPrefix string) *ProfileHandler {
	return &ProfileHandler{
		uc:           uc,
		rs:           http_response.NewResponseSender(log),
		profilesPath: strings.TrimRight(apiPrefix, "/") + "/profiles/",
	}
}

func NewProfileRouter(
	mux *http.ServeMux,
	db repository.PgxIface,
	apiPrefix string,
	appLog *logger.Logger,
	uploadPath string,
	baseURL string,
) {
	profileRepo := repository.NewProfileRepoPostgres(db, appLog)
	profileUC := usecase.NewProfileUsecase(profileRepo)
	avatarUC := usecase.NewAvatarUsecase(profileRepo, baseURL, uploadPath)

	profileHandler := NewProfileHandler(profileUC, appLog, apiPrefix)
	avatarHandler := NewAvatarHandler(avatarUC, appLog)

	chain := func(h http.Handler) http.Handler {
		return middlewares.AccessLog(appLog,
			middlewares.CorsMiddleware( // добавлен CORS
				middlewares.CSRFTokenMiddleware( // выдаём CSRF токен (cookie) на GET
					middlewares.JWT()( // сначала аутентификация
						middlewares.CSRFMiddleware(h), // затем проверка CSRF на state-changing
					),
				),
			),
		)
	}

	authChain := func(h http.Handler) http.Handler {
		return chain(h)
	}

	mux.Handle(apiPrefix+"/profiles/",
		authChain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimRight(r.URL.Path, "/")
			if strings.HasSuffix(path, "/avatar") {
				avatarHandler.UploadAvatar(w, r)
				return
			}
			profileHandler.handleProfileRoutes(w, r)
		})),
	)
}

func (h *ProfileHandler) extractIDFromRequest(r *http.Request) string {

	path := strings.TrimPrefix(r.URL.Path, h.profilesPath)
	if path == "" {
		return ""
	}
	if i := strings.IndexByte(path, '/'); i >= 0 {
		return path[:i]
	}
	return path
}

func (h *ProfileHandler) handleProfileRoutes(w http.ResponseWriter, r *http.Request) {

	id := h.extractIDFromRequest(r)

	if id == "me" {
		if sub, _ := r.Context().Value(middlewares.CtxUserID).(string); sub != "" {
			id = sub
		} else {
			h.rs.Error(r.Context(), w, http.StatusUnauthorized, "ProfileRoutes", domain.ErrUnauthorized, nil)
			return
		}
	}

	if id == "" {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "ProfileRoutes", domain.ErrRequestParams, nil)
		return
	}

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

	sub, _ := r.Context().Value(middlewares.CtxUserID).(string)
	if sub == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "GetProfile", domain.ErrUnauthorized, nil)
		return
	}
	if sub != id {
		h.rs.Error(r.Context(), w, http.StatusForbidden, "GetProfile", domain.ErrForbidden, nil)
		return
	}

	profile, err := h.uc.GetProfile(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetProfile", err, nil)
			return
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(r.Context(), w, http.StatusNotFound, "GetProfile", err, nil)
			return
		default:
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetProfile", domain.ErrInternalServer, err)
			return
		}
	}

	responseProfile := transport.ToProfileResponse(profile)
	h.rs.Send(r.Context(), w, http.StatusOK, responseProfile)
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
	sub, _ := r.Context().Value(middlewares.CtxUserID).(string)
	if sub == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "UpdateProfile", domain.ErrUnauthorized, nil)
		return
	}

	targetID := sub

	var req transport.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateProfile", domain.ErrRequestParams, err)
		return
	}

	profile := &domain.Profile{
		ID:      targetID,
		Name:    req.Name,
		Phone:   req.Phone,
		CityID:  req.CityID,
		Address: req.Address,
	}

	err := h.uc.UpdateProfile(r.Context(), profile)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "UpdateProfile", err, nil)
			return
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(r.Context(), w, http.StatusNotFound, "UpdateProfile", err, nil)
			return
		default:
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "UpdateProfile", domain.ErrInternalServer, err)
			return
		}
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
	sub, _ := r.Context().Value(middlewares.CtxUserID).(string)
	if sub == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "DeleteProfile", domain.ErrUnauthorized, nil)
		return
	}
	if sub != id {
		h.rs.Error(r.Context(), w, http.StatusForbidden, "DeleteProfile", domain.ErrForbidden, nil)
		return
	}

	err := h.uc.DeleteProfile(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "DeleteProfile", err, nil)
			return
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(r.Context(), w, http.StatusNotFound, "DeleteProfile", err, nil)
			return
		default:
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "DeleteProfile", domain.ErrInternalServer, err)
			return
		}
	}

	h.rs.Send(r.Context(), w, http.StatusNoContent, nil)
}
