package domain

import "time"

type Promocode struct {
	ID               string
	Code             string
	RelativeDiscount float64 // %
	AbsoluteDiscount float64 // рублей
	StartAt          time.Time
	EndAt            time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PromoCheckResult struct {
	RelativeDiscount float64
	AbsoluteDiscount float64
}
