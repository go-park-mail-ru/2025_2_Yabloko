package domain

import "time"

type Profile struct {
	ID        string
	Email     string
	Name      *string
	Phone     *string
	CityID    *string
	Address   *string
	AvatarURL *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
