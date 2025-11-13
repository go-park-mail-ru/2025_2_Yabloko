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
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

// derefString безопасно разыменовывает *string, возвращая "<nil>", если указатель nil
func derefString(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func NewProfileHandler(uc ProfileUsecaseInterface, apiPrefix string) *ProfileHandler {
	return &ProfileHandler{
		uc:           uc,
		rs:           http_response.NewResponseSender(logger.Global()),
		profilesPath: strings.TrimRight(apiPrefix, "/") + "/profiles/",
	}
}

// ServeAvatarStatic обрабатывает статические файлы аватарок
func ServeAvatarStatic(uploadPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := filepath.Base(r.URL.Path)
		if name == "." || name == ".." || strings.Contains(name, "/") {
			http.NotFound(w, r)
			return
		}

		if !strings.Contains(name, "_") {
			http.NotFound(w, r)
			return
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			http.NotFound(w, r)
			return
		}

		fullPath := filepath.Join(uploadPath, name)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, fullPath)
	}
}

func NewProfileRouter(
	mux *http.ServeMux,
	db repository.PgxIface,
	apiPrefix string,
	uploadPath string,
	baseURL string,
) {
	profileRepo := repository.NewProfileRepoPostgres(db)
	profileUC := usecase.NewProfileUsecase(profileRepo)
	avatarUC := usecase.NewAvatarUsecase(profileRepo, baseURL, uploadPath)

	profileHandler := NewProfileHandler(profileUC, apiPrefix)
	avatarHandler := NewAvatarHandler(avatarUC)

	mux.Handle(apiPrefix+"/profiles/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimRight(r.URL.Path, "/")
			if strings.HasSuffix(path, "/avatar") {
				avatarHandler.UploadAvatar(w, r)
				return
			}
			profileHandler.handleProfileRoutes(w, r)
		}),
	)

	// mux.HandleFunc("/", ServeAvatarStatic(uploadPath))
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler handleProfileRoutes start")

	id := h.extractIDFromRequest(r)

	if id == "me" {
		if sub, ok := middlewares.UserIDFromContext(ctx); ok && sub != "" {
			id = sub
			log.DebugContext(ctx, "handler handleProfileRoutes resolved 'me'", slog.String("user_id", id))
		} else {
			log.WarnContext(ctx, "handler handleProfileRoutes unauthorized - no user in context for 'me'")
			h.rs.Error(ctx, w, http.StatusUnauthorized, "ProfileRoutes", domain.ErrUnauthorized, nil)
			return
		}
	}

	if id == "" {
		log.WarnContext(ctx, "handler handleProfileRoutes empty id")
		h.rs.Error(ctx, w, http.StatusBadRequest, "ProfileRoutes", domain.ErrRequestParams, nil)
		return
	}

	log.InfoContext(ctx, "handler handleProfileRoutes routing",
		slog.String("user_id", id),
		slog.String("method", r.Method))

	switch r.Method {
	case http.MethodGet:
		h.GetProfile(w, r, id)
	case http.MethodPut:
		h.UpdateProfile(w, r, id)
	case http.MethodDelete:
		h.DeleteProfile(w, r, id)
	default:
		log.WarnContext(ctx, "handler handleProfileRoutes method not allowed", slog.String("method", r.Method))
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "ProfileRoutes", domain.ErrHTTPMethod, nil)
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler GetProfile start", slog.String("user_id", id))

	sub, ok := middlewares.UserIDFromContext(ctx)
	if !ok || sub == "" {
		log.WarnContext(ctx, "handler GetProfile unauthorized - no user in context")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetProfile", domain.ErrUnauthorized, nil)
		return
	}
	if sub != id {
		log.WarnContext(ctx, "handler GetProfile forbidden",
			slog.String("subject", sub),
			slog.String("target", id))
		h.rs.Error(ctx, w, http.StatusForbidden, "GetProfile", domain.ErrForbidden, nil)
		return
	}

	profile, err := h.uc.GetProfile(ctx, id)
	if err != nil {
		log.ErrorContext(ctx, "handler GetProfile usecase failed",
			slog.Any("err", err),
			slog.String("user_id", id))
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(ctx, w, http.StatusBadRequest, "GetProfile", err, nil)
			return
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "GetProfile", err, nil)
			return
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetProfile", domain.ErrInternalServer, err)
			return
		}
	}

	log.InfoContext(ctx, "handler GetProfile success",
		slog.String("user_id", id),
		slog.String("name", derefString(profile.Name)))
	responseProfile := transport.ToProfileResponse(profile)
	h.rs.Send(ctx, w, http.StatusOK, responseProfile)
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler UpdateProfile start", slog.String("user_id", id))

	sub, ok := middlewares.UserIDFromContext(ctx)
	if !ok || sub == "" {
		log.WarnContext(ctx, "handler UpdateProfile unauthorized - no user in context")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "UpdateProfile", domain.ErrUnauthorized, nil)
		return
	}
	targetID := sub

	var req transport.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "handler UpdateProfile decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateProfile", domain.ErrRequestParams, err)
		return
	}

	log.InfoContext(ctx, "handler UpdateProfile processing",
		slog.String("user_id", targetID),
		slog.String("name", derefString(req.Name)),
		slog.String("city_id", derefString(req.CityID)))

	profile := &domain.Profile{
		ID:      targetID,
		Name:    req.Name,
		Phone:   req.Phone,
		CityID:  req.CityID,
		Address: req.Address,
	}

	err := h.uc.UpdateProfile(ctx, profile)
	if err != nil {
		log.ErrorContext(ctx, "handler UpdateProfile usecase failed",
			slog.Any("err", err),
			slog.String("user_id", targetID))
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateProfile", err, nil)
			return
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "UpdateProfile", err, nil)
			return
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "UpdateProfile", domain.ErrInternalServer, err)
			return
		}
	}

	log.InfoContext(ctx, "handler UpdateProfile success", slog.String("user_id", targetID))
	h.rs.Send(ctx, w, http.StatusOK, map[string]string{"message": "Профиль успешно обновлен"})
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler DeleteProfile start", slog.String("user_id", id))

	sub, ok := middlewares.UserIDFromContext(ctx)
	if !ok || sub == "" {
		log.WarnContext(ctx, "handler DeleteProfile unauthorized - no user in context")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "DeleteProfile", domain.ErrUnauthorized, nil)
		return
	}
	if sub != id {
		log.WarnContext(ctx, "handler DeleteProfile forbidden",
			slog.String("subject", sub),
			slog.String("target", id))
		h.rs.Error(ctx, w, http.StatusForbidden, "DeleteProfile", domain.ErrForbidden, nil)
		return
	}

	err := h.uc.DeleteProfile(ctx, id)
	if err != nil {
		log.ErrorContext(ctx, "handler DeleteProfile usecase failed",
			slog.Any("err", err),
			slog.String("user_id", id))
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(ctx, w, http.StatusBadRequest, "DeleteProfile", err, nil)
			return
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "DeleteProfile", err, nil)
			return
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "DeleteProfile", domain.ErrInternalServer, err)
			return
		}
	}

	log.InfoContext(ctx, "handler DeleteProfile success", slog.String("user_id", id))
	h.rs.Send(ctx, w, http.StatusNoContent, nil)
}
