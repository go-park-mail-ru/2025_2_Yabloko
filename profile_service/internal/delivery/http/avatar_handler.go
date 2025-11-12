package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/delivery/middlewares"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"
	"io"
	"log/slog"
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

func NewAvatarHandler(avatarUC AvatarUsecaseInterface) *AvatarHandler {
	return &AvatarHandler{
		avatarUC: avatarUC,
		rs:       http_response.NewResponseSender(logger.Global()),
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler UploadAvatar start")

	if r.Method != http.MethodPost {
		log.WarnContext(ctx, "handler UploadAvatar wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "UploadAvatar", domain.ErrHTTPMethod, nil)
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[len(parts)-1] != "avatar" {
		log.WarnContext(ctx, "handler UploadAvatar invalid path", slog.String("path", r.URL.Path))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UploadAvatar", domain.ErrRequestParams, nil)
		return
	}
	userID := parts[len(parts)-2]

	subject, ok := middlewares.UserIDFromContext(ctx)
	if !ok || subject == "" {
		log.WarnContext(ctx, "handler UploadAvatar unauthorized - no user in context")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "UploadAvatar", domain.ErrUnauthorized, nil)
		return
	}

	if userID == "me" {
		userID = subject
		log.DebugContext(ctx, "handler UploadAvatar resolved 'me'", slog.String("user_id", userID))
	}

	if subject != userID {
		log.WarnContext(ctx, "handler UploadAvatar forbidden",
			slog.String("subject", subject),
			slog.String("target", userID))
		h.rs.Error(ctx, w, http.StatusForbidden, "UploadAvatar", domain.ErrForbidden, nil)
		return
	}

	const maxUpload = 10 << 20 // 10 MiB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(maxUpload); err != nil {
		log.ErrorContext(ctx, "handler UploadAvatar parse multipart failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusRequestEntityTooLarge, "UploadAvatar", domain.ErrRequestParams, err)
		return
	}

	file, fh, err := r.FormFile("avatar")
	if err != nil {
		log.ErrorContext(ctx, "handler UploadAvatar get form file failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UploadAvatar", domain.ErrRequestParams, err)
		return
	}
	defer file.Close()

	log.InfoContext(ctx, "handler UploadAvatar processing file",
		slog.String("filename", fh.Filename),
		slog.Int64("size", fh.Size),
		slog.String("user_id", userID))

	url, err := h.avatarUC.UploadAvatar(ctx, userID, file, fh)
	if err != nil {
		log.ErrorContext(ctx, "handler UploadAvatar usecase failed",
			slog.Any("err", err),
			slog.String("user_id", userID))
		switch {
		case errors.Is(err, domain.ErrInvalidProfileData):
			h.rs.Error(ctx, w, http.StatusBadRequest, "UploadAvatar", err, nil)
		case errors.Is(err, domain.ErrProfileNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "UploadAvatar", err, nil)
		case errors.Is(err, domain.ErrInvalidFileType):
			h.rs.Error(ctx, w, http.StatusUnsupportedMediaType, "UploadAvatar", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "UploadAvatar", domain.ErrInternalServer, err)
		}
		return
	}

	log.InfoContext(ctx, "handler UploadAvatar success",
		slog.String("user_id", userID),
		slog.String("avatar_url", url))
	h.rs.Send(ctx, w, http.StatusOK, map[string]string{"avatar_url": url})
}
