package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"apple_backend/profile_service/internal/domain"

	"github.com/google/uuid"
)

type AvatarUsecase struct {
	repo       ProfileRepository
	baseURL    string
	uploadPath string
}

func NewAvatarUsecase(repo ProfileRepository, baseURL, uploadPath string) *AvatarUsecase {
	return &AvatarUsecase{repo: repo, baseURL: strings.TrimRight(baseURL, "/"), uploadPath: uploadPath}
}

var allowedMIMEs = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/webp": ".webp",
}

func (uc *AvatarUsecase) UploadAvatar(ctx context.Context, userID string, src io.Reader, _ *multipart.FileHeader) (string, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return "", domain.ErrInvalidProfileData
	}

	head := make([]byte, 512)
	n, readErr := io.ReadFull(src, head)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return "", domain.ErrInvalidFileType
	}
	if n == 0 {
		return "", domain.ErrInvalidFileType
	}
	mime := http.DetectContentType(head[:n])
	ext, ok := allowedMIMEs[mime]
	if !ok {
		return "", domain.ErrInvalidFileType
	}

	var reader io.Reader
	if seeker, ok := src.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
		reader = src
	} else {
		reader = io.MultiReader(bytes.NewReader(head[:n]), src)
	}

	filename := userID + "_" + time.Now().UTC().Format("20060102T150405.000Z0700") + ext
	if err := os.MkdirAll(uc.uploadPath, 0o755); err != nil {
		return "", err
	}
	dstPath := filepath.Join(uc.uploadPath, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(dst, reader); err != nil {
		dst.Close()
		_ = os.Remove(dstPath)
		return "", err
	}
	_ = dst.Close()

	profile, err := uc.repo.GetProfile(ctx, userID)
	if err != nil {
		_ = os.Remove(dstPath)
		return "", err
	}
	if profile.AvatarURL != nil && *profile.AvatarURL != "" {
		if u, perr := url.Parse(*profile.AvatarURL); perr == nil {
			oldName := filepath.Base(u.Path)
			if oldName != "" {
				_ = os.Remove(filepath.Join(uc.uploadPath, oldName))
			}
		}
	}

	avatarURL := uc.baseURL + "/" + filename
	profile.AvatarURL = &avatarURL

	if err := uc.repo.UpdateProfile(ctx, profile); err != nil {
		_ = os.Remove(dstPath)
		return "", err
	}
	return avatarURL, nil
}
