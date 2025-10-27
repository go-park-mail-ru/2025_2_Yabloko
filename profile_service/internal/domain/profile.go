package domain

import "time"

type Profile struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         *string   `json:"name,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	CityID       *string   `json:"city_id,omitempty"`
	Address      *string   `json:"address,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
