package transport

import (
	"apple_backend/support_service/internal/domain"
	"time"
)

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

type MessageResponse struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	UserID    *string   `json:"user_id,omitempty"`
	GuestID   *string   `json:"guest_id,omitempty"`
	UserRole  string    `json:"user_role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type RatingResponse struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketWithMessagesResponse struct {
	Ticket   *TicketResponse    `json:"ticket"`
	Messages []*MessageResponse `json:"messages"`
	Rating   *RatingResponse    `json:"rating,omitempty"`
}

type StatisticsResponse struct {
	TotalTickets      int            `json:"total_tickets"`
	OpenTickets       int            `json:"open_tickets"`
	InProgressTickets int            `json:"in_progress_tickets"`
	ClosedTickets     int            `json:"closed_tickets"`
	TicketsByCategory map[string]int `json:"tickets_by_category"`
	AverageRating     float64        `json:"average_rating"`
}

func ToTicketResponse(t *domain.Ticket) *TicketResponse {
	if t == nil {
		return nil
	}
	return &TicketResponse{
		ID:          t.ID,
		UserName:    t.UserName,
		UserEmail:   t.UserEmail,
		Category:    t.Category,
		Status:      t.Status,
		Priority:    t.Priority,
		Title:       t.Title,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToTicketResponses(tickets []*domain.Ticket) []*TicketResponse {
	responses := make([]*TicketResponse, len(tickets))
	for i, t := range tickets {
		responses[i] = ToTicketResponse(t)
	}
	return responses
}

func ToMessageResponse(m *domain.Message) *MessageResponse {
	if m == nil {
		return nil
	}
	return &MessageResponse{
		ID:        m.ID,
		TicketID:  m.TicketID,
		UserID:    m.UserID,
		GuestID:   m.GuestID,
		UserRole:  m.UserRole,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func ToMessageResponses(messages []*domain.Message) []*MessageResponse {
	responses := make([]*MessageResponse, len(messages))
	for i, m := range messages {
		responses[i] = ToMessageResponse(m)
	}
	return responses
}

func ToRatingResponse(r *domain.Rating) *RatingResponse {
	if r == nil {
		return nil
	}
	var comment *string
	if r.Comment != "" {
		comment = &r.Comment
	}
	return &RatingResponse{
		ID:        r.ID,
		TicketID:  r.TicketID,
		Rating:    r.Rating,
		Comment:   comment,
		CreatedAt: r.CreatedAt,
	}
}

func ToTicketWithMessagesResponse(twm *domain.TicketWithMessages) *TicketWithMessagesResponse {
	if twm == nil {
		return nil
	}
	return &TicketWithMessagesResponse{
		Ticket:   ToTicketResponse(twm.Ticket),
		Messages: ToMessageResponses(twm.Messages),
		Rating:   ToRatingResponse(twm.Rating),
	}
}

func ToStatisticsResponse(s *domain.Statistics) *StatisticsResponse {
	if s == nil {
		return nil
	}
	return &StatisticsResponse{
		TotalTickets:      s.TotalTickets,
		OpenTickets:       s.OpenTickets,
		InProgressTickets: s.InProgressTickets,
		ClosedTickets:     s.ClosedTickets,
		TicketsByCategory: s.TicketsByCategory,
		AverageRating:     s.AverageRating,
	}
}

type CreateTicketRequest struct {
	UserName    string `json:"user_name,omitempty" validate:"omitempty,max=100"`
	UserEmail   string `json:"user_email,omitempty" validate:"omitempty,email,max=100"`
	Category    string `json:"category" validate:"required,oneof=bug feature complaint"`
	Title       string `json:"title" validate:"required,max=200"`
	Description string `json:"description" validate:"required,max=2000"`
}

type AddMessageRequest struct {
	Content string `json:"content" validate:"required,max=2000"`
}

type UpdateTicketRequest struct {
	Title       string `json:"title" validate:"required,max=200"`
	Description string `json:"description" validate:"required,max=2000"`
	Category    string `json:"category" validate:"required,oneof=bug feature complaint"`
}

// AddRatingRequest - DTO для добавления оценки
type AddRatingRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment,omitempty" validate:"omitempty,max=500"`
}

// UpdateStatusRequest - DTO для обновления статуса (админ)
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=open in_progress closed"`
}
