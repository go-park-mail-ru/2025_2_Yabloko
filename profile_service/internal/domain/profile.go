package domain

import "time"

type Profile struct {
	ID           string
	Email        string
	PasswordHash string
	Name         *string
	Phone        *string
	CityID       *string
	Address      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
