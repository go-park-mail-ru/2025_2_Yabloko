package domain

import "errors"

var (
	ErrRowsNotFound    = errors.New("rows not found")
	ErrRequestParams   = errors.New("invalid request parameters")
	ErrHTTPMethod      = errors.New("http method not allowed")
	ErrInternalServer  = errors.New("internal server error")
	ErrAccessDenied    = errors.New("access denied")
	ErrTicketNotClosed = errors.New("ticket must be closed to add rating")
	ErrRatingExists    = errors.New("rating already exists for this ticket")
	ErrUserNotFound    = errors.New("user not found")
	ErrTicketClosed    = errors.New("ticket is closed")
)
