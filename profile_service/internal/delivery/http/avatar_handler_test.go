package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"

	"log/slog"

	"github.com/stretchr/testify/require"
)

type mockAvatarUC struct {
	UploadAvatarFunc func(ctx context.Context, userID string, file io.Reader, fh *multipart.FileHeader) (string, error)
}

func (m *mockAvatarUC) UploadAvatar(ctx context.Context, userID string, file io.Reader, fh *multipart.FileHeader) (string, error) {
	return m.UploadAvatarFunc(ctx, userID, file, fh)
}

func TestAvatarHandler_UploadAvatar(t *testing.T) {
	log := logger.NewLogger("test", slog.LevelError)
	handler := NewAvatarHandler(nil, log)

	userID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("Неверный метод", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profiles/"+userID+"/avatar", nil)
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Неправильный путь", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/wrong", nil)
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Путь с завершающим слешем", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		_, _ = writer.CreateFormFile("avatar", "avatar.jpg")
		_ = writer.Close()

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "http://localhost/avatar.jpg", nil
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar/", buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Ошибка FormFile (нет части avatar)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", nil)
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Usecase ErrInvalidProfileData", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		_, _ = writer.CreateFormFile("avatar", "avatar.jpg")
		_ = writer.Close()

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "", domain.ErrInvalidProfileData
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Usecase ErrProfileNotFound", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		_, _ = writer.CreateFormFile("avatar", "avatar.jpg")
		_ = writer.Close()

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "", domain.ErrProfileNotFound
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Usecase ErrInvalidFileType", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		_, _ = writer.CreateFormFile("avatar", "avatar.jpg")
		_ = writer.Close()

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "", domain.ErrInvalidFileType
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
	})

	t.Run("Usecase internal error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		_, _ = writer.CreateFormFile("avatar", "avatar.jpg")
		_ = writer.Close()

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "", errors.New("some error")
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Успешная загрузка", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		_, _ = writer.CreateFormFile("avatar", "avatar.jpg")
		_ = writer.Close()

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "http://localhost/avatar.jpg", nil
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), "http://localhost/avatar.jpg")
	})

	t.Run("Слишком большой файл → 413", func(t *testing.T) {
		var big bytes.Buffer
		writer := multipart.NewWriter(&big)

		part, err := writer.CreateFormFile("avatar", "big.jpg")
		require.NoError(t, err)

		_, _ = part.Write(bytes.Repeat([]byte{1}, (10<<20)+(1<<10)))
		require.NoError(t, writer.Close())

		handler.avatarUC = &mockAvatarUC{
			UploadAvatarFunc: func(ctx context.Context, uid string, file io.Reader, fh *multipart.FileHeader) (string, error) {
				return "won't be called", nil
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/profiles/"+userID+"/avatar", &big)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler.UploadAvatar(rr, req)
		require.Equal(t, http.StatusRequestEntityTooLarge, rr.Code)
	})
}
