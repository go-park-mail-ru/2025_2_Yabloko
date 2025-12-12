package transport

type PromoCheckRequest struct {
	Promo string `json:"promo" validate:"required,min=3,max=32"`
} // @name PromoCheckRequest

type PromoCheckResponse struct {
	RelativeDiscount float64 `json:"relative_discount"` // %
	AbsoluteDiscount float64 `json:"absolute_discount"` // Рублей
} // @name PromoCheckResponse
