package transport

import (
	"support_service/internal/domain"
	"time"
)

// Request DTOs
type CreateTicketRequest struct {
	UserName    string `json:"user_name"`
	UserEmail   string `json:"user_email"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AddMessageRequest struct {
	Content string `json:"content"`
}

// Response DTOs
type TicketResponse struct {
	ID          string    `json:"id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Конвертеры domain -> transport
func ToTicketResponse(ticket *domain.Ticket) *TicketResponse {
	if ticket == nil {
		return nil
	}
	return &TicketResponse{
		ID:          ticket.ID,
		UserName:    ticket.UserName,
		UserEmail:   ticket.UserEmail,
		Category:    ticket.Category,
		Status:      ticket.Status,
		Priority:    ticket.Priority,
		Title:       ticket.Title,
		Description: ticket.Description,
		CreatedAt:   ticket.CreatedAt,
		UpdatedAt:   ticket.UpdatedAt,
	}
}
