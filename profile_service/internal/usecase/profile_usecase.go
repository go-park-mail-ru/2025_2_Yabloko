//go:generate mockgen -source=profile_usecase.go -destination=mock/profile_repository_mock.go -package=mock
package usecase

import (
	"context"
	"strings"

	"apple_backend/profile_service/internal/domain"

	"github.com/google/uuid"
)

type ProfileUsecase struct {
	repo ProfileRepository
}

func NewProfileUsecase(repo ProfileRepository) *ProfileUsecase {
	return &ProfileUsecase{repo: repo}
}

func (uc *ProfileUsecase) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.ErrInvalidProfileData
	}
	return uc.repo.GetProfile(ctx, id)
}

func (uc *ProfileUsecase) UpdateProfile(ctx context.Context, in *domain.Profile) error {
	if _, err := uuid.Parse(in.ID); err != nil {
		return domain.ErrInvalidProfileData
	}
	existing, err := uc.repo.GetProfile(ctx, in.ID)
	if err != nil {
		return err
	}

	if in.Name != nil && len(*in.Name) > 100 {
		return domain.ErrInvalidProfileData
	}
	if in.Phone != nil {
		p := strings.TrimSpace(*in.Phone)
		if len(p) < 10 || len(p) > 20 {
			return domain.ErrInvalidProfileData
		}
	}
	if in.Address != nil && len(*in.Address) > 200 {
		return domain.ErrInvalidProfileData
	}
	if in.Name != nil {
		existing.Name = in.Name
	}
	if in.Phone != nil {
		existing.Phone = in.Phone
	}
	if in.CityID != nil {
		existing.CityID = in.CityID
	}
	if in.Address != nil {
		existing.Address = in.Address
	}

	return uc.repo.UpdateProfile(ctx, existing)
}

func (uc *ProfileUsecase) DeleteProfile(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return domain.ErrInvalidProfileData
	}
	return uc.repo.DeleteProfile(ctx, id)
}
