package transport

import (
	"apple_backend/profile_service/internal/domain"
	"time"
)

type CreateProfileResponse struct {
	ID string `json:"id"`
} // @name CreateProfileResponse

type ProfileResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      *string `json:"name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	CityID    *string `json:"city_id,omitempty"`
	Address   *string `json:"address,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
} // @name ProfileResponse

type CreateProfileRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} // @name CreateProfileRequest

type UpdateProfileRequest struct {
	Name    *string `json:"name,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	CityID  *string `json:"city_id,omitempty"`
	Address *string `json:"address,omitempty"`
} // @name UpdateProfileRequest

func ToProfileResponse(p *domain.Profile) *ProfileResponse {
	return &ProfileResponse{
		ID:        p.ID,
		Email:     p.Email,
		Name:      p.Name,
		Phone:     p.Phone,
		CityID:    p.CityID,
		Address:   p.Address,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}
