package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"context"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id string) (*domain.Profile, error)
	GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, profile *domain.Profile) error
	DeleteProfile(ctx context.Context, id string) error
}

type ProfileUsecase struct {
	repo ProfileRepository
}

func NewProfileUsecase(repo ProfileRepository) *ProfileUsecase {
	return &ProfileUsecase{repo: repo}
}

func (uc *ProfileUsecase) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	return uc.repo.GetProfile(ctx, id)
}

func (uc *ProfileUsecase) GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error) {
	return uc.repo.GetProfileByEmail(ctx, email)
}

func (uc *ProfileUsecase) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	return uc.repo.UpdateProfile(ctx, profile)
}

func (uc *ProfileUsecase) DeleteProfile(ctx context.Context, id string) error {
	return uc.repo.DeleteProfile(ctx, id)
}
