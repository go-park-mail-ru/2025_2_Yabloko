//go:generate mockgen -source=profile_usecase.go -destination=mock/profile_repository_mock.go -package=mock
package usecase

import (
	"apple_backend/profile_service/internal/domain"
	"context"
	"strings"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id string) (*domain.Profile, error)
	GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error)
	CreateProfile(ctx context.Context, profile *domain.Profile) error
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

func (uc *ProfileUsecase) CreateProfile(ctx context.Context, email, passwordHash string) (*domain.Profile, error) {
	_, err := uc.repo.GetProfileByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrProfileExist
	}
	if err != domain.ErrProfileNotFound {
		return nil, err
	}

	profile := &domain.Profile{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := uc.repo.CreateProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (uc *ProfileUsecase) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	existing, err := uc.repo.GetProfile(ctx, profile.ID)
	if err != nil {
		return err
	}

	if profile.Name != nil && len(*profile.Name) > 100 {
		return domain.ErrInvalidProfileData
	}

	if profile.Phone != nil {
		p := strings.TrimSpace(*profile.Phone)
		if len(p) < 10 || len(p) > 20 {
			return domain.ErrInvalidProfileData
		}
	}

	if profile.Address != nil && len(*profile.Address) > 200 {
		return domain.ErrInvalidProfileData
	}

	merged := &domain.Profile{
		ID:      existing.ID,
		Email:   existing.Email,
		Name:    existing.Name,
		Phone:   existing.Phone,
		CityID:  existing.CityID,
		Address: existing.Address,
	}

	if profile.Name != nil {
		merged.Name = profile.Name
	}
	if profile.Phone != nil {
		merged.Phone = profile.Phone
	}
	if profile.CityID != nil {
		merged.CityID = profile.CityID
	}
	if profile.Address != nil {
		merged.Address = profile.Address
	}

	return uc.repo.UpdateProfile(ctx, merged)
}

func (uc *ProfileUsecase) DeleteProfile(ctx context.Context, id string) error {
	return uc.repo.DeleteProfile(ctx, id)
}
