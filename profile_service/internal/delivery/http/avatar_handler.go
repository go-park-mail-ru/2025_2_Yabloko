package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type AvatarUsecaseInterface interface {
	UploadAvatar(ctx context.Context, userID string, file io.Reader, fileHeader *multipart.FileHeader) (string, error)
}

type AvatarHandler struct {
	avatarUC AvatarUsecaseInterface
	rs       *http_response.ResponseSender
}

func NewAvatarHandler(avatarUC AvatarUsecaseInterface, log *logger.Logger) *AvatarHandler {
	return &AvatarHandler{
		avatarUC: avatarUC,
		rs:       http_response.NewResponseSender(log),
	}
}

// UploadAvatar godoc
// @Summary Загрузить аватарку
// @Description Загружает аватарку для профиля пользователя
// @Tags profiles
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "UUID пользователя"
// @Param avatar formData file true "Файл аватарки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} http_response.ErrResponse "Ошибка входных данных или файла"
// @Failure 404 {object} http_response.ErrResponse "Профиль не найден"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /profiles/{id}/avatar [post]
func (h *AvatarHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "UploadAvatar", domain.ErrHTTPMethod, nil)
		return
	}

	// поддержка .../avatar и .../avatar/
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[len(parts)-1] != "avatar" {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UploadAvatar", domain.ErrRequestParams, nil)
		return
	}
	userID := parts[len(parts)-2]

	const maxUpload = 10 << 20 // 10 MiB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(maxUpload); err != nil {
		h.rs.Error(r.Context(), w, http.StatusRequestEntityTooLarge, "UploadAvatar", domain.ErrRequestParams, err)
		return
	}

	file, fh, err := r.FormFile("avatar")
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "UploadAvatar", domain.ErrRequestParams, err)
		return
	}
	defer file.Close()

	url, err := h.avatarUC.UploadAvatar(r.Context(), userID, file, fh)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "UploadAvatar", err, nil)
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(r.Context(), w, http.StatusNotFound, "UploadAvatar", err, nil)
		case errors.Is(err, domain.ErrInvalidFileType):
			h.rs.Error(r.Context(), w, http.StatusUnsupportedMediaType, "UploadAvatar", err, nil)
		default:
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "UploadAvatar", domain.ErrInternalServer, err)
		}
		return
	}

	h.rs.Send(r.Context(), w, http.StatusOK, map[string]string{"avatar_url": url})
}
