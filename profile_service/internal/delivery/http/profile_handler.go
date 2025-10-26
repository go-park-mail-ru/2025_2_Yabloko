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

	"github.com/google/uuid"
)

type ProfileUsecaseInterface interface {
	GetProfile(ctx context.Context, id string) (*domain.Profile, error)
	GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error)
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

func NewProfileRouter(db repository.PgxIface, apiPrefix string, appLog, accessLog *logger.Logger) http.Handler {
	profileRepo := repository.NewProfileRepoPostgres(db, appLog)
	profileUC := usecase.NewProfileUsecase(profileRepo)
	profileHandler := NewProfileHandler(profileUC, appLog)

	mux := http.NewServeMux()

	mux.HandleFunc(apiPrefix+"/profiles/", profileHandler.handleProfileRoutes)
	mux.HandleFunc(apiPrefix+"/profiles/email/", profileHandler.GetProfileByEmail)

	return middlewares.CSRFMiddleware(
		middlewares.CSRFTokenMiddleware(
			middlewares.CorsMiddleware(
				middlewares.AccessLog(accessLog, mux),
			),
		),
	)
}

func extractIDFromPath(path, prefix string) string {
	id := strings.TrimPrefix(path, prefix)
	return strings.TrimSuffix(id, "/")
}

func (h *ProfileHandler) handleProfileRoutes(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v0/profiles/")

	if _, err := uuid.Parse(id); err != nil && id != "" {
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
