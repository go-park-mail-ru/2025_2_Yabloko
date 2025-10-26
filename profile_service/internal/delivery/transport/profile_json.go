package transport

import (
	"apple_backend/profile_service/internal/domain"
	"html"
)

type ProfileResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      *string `json:"name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	CityID    *string `json:"city_id,omitempty"`
	Address   *string `json:"address,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type UpdateProfileRequest struct {
	Name    *string `json:"name,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	CityID  *string `json:"city_id,omitempty"`
	Address *string `json:"address,omitempty"`
}

func escapeString(s *string) *string {
	if s == nil {
		return nil
	}
	escaped := html.EscapeString(*s)
	return &escaped
}

func ToProfileResponse(profile *domain.Profile) *ProfileResponse {
	if profile == nil {
		return nil
	}

	return &ProfileResponse{
		ID:        profile.ID,
		Email:     profile.Email,
		Name:      escapeString(profile.Name),
		Phone:     escapeString(profile.Phone),
		CityID:    escapeString(profile.CityID),
		Address:   escapeString(profile.Address),
		CreatedAt: profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
