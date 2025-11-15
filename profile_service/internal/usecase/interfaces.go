package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"context"
	"io"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id string) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, profile *domain.Profile) error
	DeleteProfile(ctx context.Context, id string) error
}

// пока что для будущего перехода на s3
type AvatarStorage interface {
	Upload(ctx context.Context, key string, file io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
}
