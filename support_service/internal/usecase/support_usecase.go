package usecase

import (
	"apple_backend/support_service/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type TicketRepository interface {
	CreateTicket(ctx context.Context, ticket *domain.Ticket) error
	GetTicket(ctx context.Context, id string) (*domain.Ticket, error)
	GetUserTickets(ctx context.Context, userID, guestID *string) ([]*domain.Ticket, error)
	GetAllTickets(ctx context.Context, filter *domain.TicketFilter) ([]*domain.Ticket, error)
	UpdateTicketStatus(ctx context.Context, id, status string) error
	UpdateTicket(ctx context.Context, ticket *domain.Ticket) error
	GetStatistics(ctx context.Context) (*domain.Statistics, error)
	GetUserByID(ctx context.Context, userID string) (*domain.UserInfo, error)
}

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *domain.Message) error
	GetMessagesByTicket(ctx context.Context, ticketID string) ([]*domain.Message, error)
}

type RatingRepository interface {
	CreateRating(ctx context.Context, rating *domain.Rating) error
	GetRatingByTicket(ctx context.Context, ticketID string) (*domain.Rating, error)
	GetAverageRating(ctx context.Context) (float64, error)
}

type SupportUsecase struct {
	ticketRepo  TicketRepository
	messageRepo MessageRepository
	ratingRepo  RatingRepository
}

func NewSupportUsecase(
	ticketRepo TicketRepository,
	messageRepo MessageRepository,
	ratingRepo RatingRepository,
) *SupportUsecase {
	return &SupportUsecase{
		ticketRepo:  ticketRepo,
		messageRepo: messageRepo,
		ratingRepo:  ratingRepo,
	}
}

func (uc *SupportUsecase) CreateTicket(
	ctx context.Context,
	userID, guestID *string,
	userName, userEmail, category, title, description string,
) (*domain.Ticket, error) {

	if userID == nil && guestID == nil {
		return nil, domain.ErrRequestParams
	}

	var finalUserName, finalUserEmail string

	if userID != nil {
		user, err := uc.ticketRepo.GetUserByID(ctx, *userID)
		if err != nil {
			return nil, err
		}
		finalUserName = user.Name
		finalUserEmail = user.Email
	} else {
		finalUserName = userName
		finalUserEmail = userEmail
	}

	ticket := &domain.Ticket{
		ID:          uuid.New().String(),
		UserID:      userID,
		GuestID:     guestID,
		UserName:    finalUserName,
		UserEmail:   finalUserEmail,
		Category:    category,
		Status:      "open",
		Priority:    uc.determinePriority(category),
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.ticketRepo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (uc *SupportUsecase) GetUserTickets(ctx context.Context, userID, guestID *string) ([]*domain.Ticket, error) {
	tickets, err := uc.ticketRepo.GetUserTickets(ctx, userID, guestID)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

func (uc *SupportUsecase) GetTicketWithMessages(ctx context.Context, ticketID string, userID, guestID *string) (*domain.TicketWithMessages, error) {
	ticket, err := uc.validateTicketAccess(ctx, ticketID, userID, guestID)
	if err != nil {
		return nil, err
	}

	messages, err := uc.messageRepo.GetMessagesByTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	rating, _ := uc.ratingRepo.GetRatingByTicket(ctx, ticketID)

	return &domain.TicketWithMessages{
		Ticket:   ticket,
		Messages: messages,
		Rating:   rating,
	}, nil
}

func (uc *SupportUsecase) AddMessage(
	ctx context.Context,
	ticketID string,
	userID, guestID *string,
	userRole, content string,
) error {

	ticket, err := uc.validateTicketAccess(ctx, ticketID, userID, guestID)
	if err != nil {
		return err
	}

	if ticket.Status == "closed" {
		return domain.ErrTicketClosed
	}

	message := &domain.Message{
		ID:        uuid.New().String(),
		TicketID:  ticketID,
		UserID:    userID,
		GuestID:   guestID,
		UserRole:  userRole,
		Content:   content,
		CreatedAt: time.Now(),
	}

	return uc.messageRepo.CreateMessage(ctx, message)
}

func (uc *SupportUsecase) AddRating(
	ctx context.Context,
	ticketID string,
	userID, guestID *string,
	rating int,
	comment string,
) error {

	ticket, err := uc.validateTicketAccess(ctx, ticketID, userID, guestID)
	if err != nil {
		return err
	}

	if ticket.Status != "closed" {
		return domain.ErrTicketNotClosed
	}

	existingRating, _ := uc.ratingRepo.GetRatingByTicket(ctx, ticketID)
	if existingRating != nil {
		return domain.ErrRatingExists
	}

	ratingObj := &domain.Rating{
		ID:        uuid.New().String(),
		TicketID:  ticketID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	return uc.ratingRepo.CreateRating(ctx, ratingObj)
}

func (uc *SupportUsecase) GetAllTickets(ctx context.Context, filter *domain.TicketFilter) ([]*domain.Ticket, error) {
	tickets, err := uc.ticketRepo.GetAllTickets(ctx, filter)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

func (uc *SupportUsecase) UpdateTicketStatus(ctx context.Context, ticketID, status string) error {
	return uc.ticketRepo.UpdateTicketStatus(ctx, ticketID, status)
}

func (uc *SupportUsecase) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	stats, err := uc.ticketRepo.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	avgRating, err := uc.ratingRepo.GetAverageRating(ctx)
	if err == nil {
		stats.AverageRating = avgRating
	}

	return stats, nil
}

func (uc *SupportUsecase) determinePriority(category string) string {
	switch category {
	case "bug", "complaint":
		return "high"
	case "feature":
		return "medium"
	default:
		return "low"
	}
}

func (uc *SupportUsecase) isTicketOwner(ticket *domain.Ticket, userID, guestID *string) bool {
	if userID != nil && ticket.UserID != nil && *userID == *ticket.UserID {
		return true
	}
	if guestID != nil && ticket.GuestID != nil && *guestID == *ticket.GuestID {
		return true
	}
	return false
}

func (uc *SupportUsecase) validateTicketAccess(
	ctx context.Context,
	ticketID string,
	userID, guestID *string,
) (*domain.Ticket, error) {

	ticket, err := uc.ticketRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if !uc.isTicketOwner(ticket, userID, guestID) {
		return nil, domain.ErrAccessDenied
	}

	return ticket, nil
}

func (uc *SupportUsecase) UpdateTicket(
	ctx context.Context,
	ticketID string,
	userID, guestID *string,
	title, description, category string,
) error {

	ticket, err := uc.validateTicketAccess(ctx, ticketID, userID, guestID)
	if err != nil {
		return err
	}

	ticket.Title = title
	ticket.Description = description
	ticket.Category = category
	ticket.UpdatedAt = time.Now()

	return uc.ticketRepo.UpdateTicket(ctx, ticket)
}
