package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/support_service/internal/delivery/middlewares"
	"apple_backend/support_service/internal/delivery/transport"
	"apple_backend/support_service/internal/domain"
	"apple_backend/support_service/internal/repository"
	"apple_backend/support_service/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type SupportUsecaseInterface interface {
	CreateTicket(
		ctx context.Context,
		userID, guestID *string,
		userName, userEmail, category, title, description string,
	) (*domain.Ticket, error)

	GetUserTickets(ctx context.Context, userID, guestID *string) ([]*domain.Ticket, error)

	GetTicketWithMessages(
		ctx context.Context,
		ticketID string,
		userID, guestID *string,
	) (*domain.TicketWithMessages, error)

	AddMessage(
		ctx context.Context,
		ticketID string,
		userID, guestID *string,
		userRole, content string,
	) error

	AddRating(
		ctx context.Context,
		ticketID string,
		userID, guestID *string,
		rating int,
		comment string,
	) error

	GetAllTickets(ctx context.Context, filter *domain.TicketFilter) ([]*domain.Ticket, error)

	UpdateTicketStatus(ctx context.Context, ticketID, status string) error

	GetStatistics(ctx context.Context) (*domain.Statistics, error)

	UpdateTicket(
		ctx context.Context,
		ticketID string,
		userID, guestID *string,
		title, description, category string,
	) error
}

type SupportHandler struct {
	uc SupportUsecaseInterface
	rs *http_response.ResponseSender
}

func NewSupportRouter(
	mux *http.ServeMux,
	db repository.PgxIface,
	apiPrefix string,
) {
	ticketRepo := repository.NewTicketRepoPostgres(db)
	messageRepo := repository.NewMessageRepoPostgres(db)
	ratingRepo := repository.NewRatingRepoPostgres(db)

	supportUC := usecase.NewSupportUsecase(ticketRepo, messageRepo, ratingRepo)
	supportHandler := NewSupportHandler(supportUC)

	basePath := strings.TrimRight(apiPrefix, "/")

	public := middlewares.AuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, basePath)

			switch {
			case path == "/tickets" && r.Method == http.MethodPost:
				supportHandler.CreateTicketHandler(w, r)
			case path == "/tickets" && r.Method == http.MethodGet:
				supportHandler.GetUserTicketsHandler(w, r)
			case strings.HasPrefix(path, "/tickets/") && r.Method == http.MethodGet:
				supportHandler.GetTicketWithMessagesHandler(w, r)
			case strings.HasPrefix(path, "/tickets/") && r.Method == http.MethodPut:
				supportHandler.UpdateTicketHandler(w, r)
			case strings.HasSuffix(path, "/messages") && r.Method == http.MethodPost:
				supportHandler.AddMessageHandler(w, r)
			case strings.HasSuffix(path, "/rating") && r.Method == http.MethodPost:
				supportHandler.AddRatingHandler(w, r)
			default:
				http.NotFound(w, r)
			}
		}),
	)

	admin := middlewares.AuthMiddleware(
		middlewares.AdminMiddleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path := strings.TrimPrefix(r.URL.Path, basePath)

				switch {
				case path == "/admin/tickets" && r.Method == http.MethodGet:
					supportHandler.GetAllTicketsHandler(w, r)
				case strings.HasSuffix(path, "/status") && r.Method == http.MethodPut:
					supportHandler.UpdateTicketStatusHandler(w, r)
				case path == "/admin/statistics" && r.Method == http.MethodGet:
					supportHandler.GetStatisticsHandler(w, r)
				default:
					http.NotFound(w, r)
				}
			}),
		),
	)

	mux.Handle(basePath+"/tickets", public)
	mux.Handle(basePath+"/tickets/", public)
	mux.Handle(basePath+"/admin/tickets", admin)
	mux.Handle(basePath+"/admin/tickets/", admin)
	mux.Handle(basePath+"/admin/statistics", admin)
}

func NewSupportHandler(uc SupportUsecaseInterface) *SupportHandler {
	return &SupportHandler{
		uc: uc,
		rs: http_response.NewResponseSender(logger.Global()),
	}
}

func (h *SupportHandler) getUserIdentifiers(ctx context.Context) (*string, *string) {
	userID, _ := middlewares.UserIDFromContext(ctx)
	guestID, _ := middlewares.GuestIDFromContext(ctx)

	var userIDPtr, guestIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}
	if guestID != "" {
		guestIDPtr = &guestID
	}

	return userIDPtr, guestIDPtr
}

func (h *SupportHandler) getUserRole(ctx context.Context) string {
	role, _ := middlewares.UserRoleFromContext(ctx)
	return role
}

func (h *SupportHandler) extractTicketID(r *http.Request) string {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 && parts[0] == "tickets" {
		return parts[1]
	}
	return ""
}

func (h *SupportHandler) validateEnum(value string, allowed []string) bool {
	for _, v := range allowed {
		if value == v {
			return true
		}
	}
	return false
}

func (h *SupportHandler) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPost {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "CreateTicket", domain.ErrHTTPMethod, nil)
		return
	}

	var req transport.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "CreateTicket decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, err)
		return
	}

	if req.Category == "" || req.Title == "" || req.Description == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, nil)
		return
	}

	if !h.validateEnum(req.Category, []string{"bug", "feature", "complaint"}) {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, nil)
		return
	}

	if len(req.Title) > 200 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, nil)
		return
	}

	if len(req.Description) > 2000 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, nil)
		return
	}

	if req.UserName != "" && len(req.UserName) > 100 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, nil)
		return
	}

	if req.UserEmail != "" && len(req.UserEmail) > 100 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", domain.ErrRequestParams, nil)
		return
	}

	userID, guestID := h.getUserIdentifiers(ctx)
	if userID == nil && guestID == nil {
		h.rs.Error(ctx, w, http.StatusUnauthorized, "CreateTicket", domain.ErrUnauthorized, nil)
		return
	}

	ticket, err := h.uc.CreateTicket(
		ctx,
		userID,
		guestID,
		req.UserName,
		req.UserEmail,
		req.Category,
		req.Title,
		req.Description,
	)

	if err != nil {
		log.ErrorContext(ctx, "CreateTicket usecase failed", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrRequestParams):
			h.rs.Error(ctx, w, http.StatusBadRequest, "CreateTicket", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "CreateTicket", domain.ErrInternalServer, err)
		}
		return
	}

	response := transport.ToTicketResponse(ticket)
	h.rs.Send(ctx, w, http.StatusCreated, response)
}

// GetUserTicketsHandler - GET /tickets
func (h *SupportHandler) GetUserTicketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodGet {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetUserTickets", domain.ErrHTTPMethod, nil)
		return
	}

	userID, guestID := h.getUserIdentifiers(ctx)
	if userID == nil && guestID == nil {
		h.rs.Error(ctx, w, http.StatusUnauthorized, "GetUserTickets", domain.ErrUnauthorized, nil)
		return
	}

	tickets, err := h.uc.GetUserTickets(ctx, userID, guestID)
	if err != nil {
		log.ErrorContext(ctx, "GetUserTickets usecase failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetUserTickets", domain.ErrInternalServer, err)
		return
	}

	response := transport.ToTicketResponses(tickets)
	h.rs.Send(ctx, w, http.StatusOK, response)
}

// GetTicketWithMessagesHandler - GET /tickets/{id}
func (h *SupportHandler) GetTicketWithMessagesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodGet {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetTicketWithMessages", domain.ErrHTTPMethod, nil)
		return
	}

	ticketID := h.extractTicketID(r)
	if ticketID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "GetTicketWithMessages", domain.ErrRequestParams, nil)
		return
	}

	userID, guestID := h.getUserIdentifiers(ctx)

	result, err := h.uc.GetTicketWithMessages(ctx, ticketID, userID, guestID)
	if err != nil {
		log.ErrorContext(ctx, "GetTicketWithMessages usecase failed", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrAccessDenied):
			h.rs.Error(ctx, w, http.StatusForbidden, "GetTicketWithMessages", err, nil)
		case errors.Is(err, domain.ErrRowsNotFound):
			h.rs.Error(ctx, w, http.StatusNotFound, "GetTicketWithMessages", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "GetTicketWithMessages", domain.ErrInternalServer, err)
		}
		return
	}

	response := transport.ToTicketWithMessagesResponse(result)
	h.rs.Send(ctx, w, http.StatusOK, response)
}

// UpdateTicketHandler - PUT /tickets/{id}
func (h *SupportHandler) UpdateTicketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPut {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "UpdateTicket", domain.ErrHTTPMethod, nil)
		return
	}

	ticketID := h.extractTicketID(r)
	if ticketID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicket", domain.ErrRequestParams, nil)
		return
	}

	var req transport.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "UpdateTicket decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicket", domain.ErrRequestParams, err)
		return
	}

	if req.Title == "" || req.Description == "" || req.Category == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicket", domain.ErrRequestParams, nil)
		return
	}

	if !h.validateEnum(req.Category, []string{"bug", "feature", "complaint"}) {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicket", domain.ErrRequestParams, nil)
		return
	}

	if len(req.Title) > 200 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicket", domain.ErrRequestParams, nil)
		return
	}

	if len(req.Description) > 2000 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicket", domain.ErrRequestParams, nil)
		return
	}

	userID, guestID := h.getUserIdentifiers(ctx)

	err := h.uc.UpdateTicket(
		ctx,
		ticketID,
		userID,
		guestID,
		req.Title,
		req.Description,
		req.Category,
	)

	if err != nil {
		log.ErrorContext(ctx, "UpdateTicket usecase failed", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrAccessDenied):
			h.rs.Error(ctx, w, http.StatusForbidden, "UpdateTicket", err, nil)
		case errors.Is(err, domain.ErrTicketClosed):
			h.rs.Error(ctx, w, http.StatusConflict, "UpdateTicket", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "UpdateTicket", domain.ErrInternalServer, err)
		}
		return
	}

	h.rs.Send(ctx, w, http.StatusNoContent, nil)
}

// AddMessageHandler - POST /tickets/{id}/messages
func (h *SupportHandler) AddMessageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPost {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "AddMessage", domain.ErrHTTPMethod, nil)
		return
	}

	ticketID := h.extractTicketID(r)
	if ticketID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddMessage", domain.ErrRequestParams, nil)
		return
	}

	var req transport.AddMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "AddMessage decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddMessage", domain.ErrRequestParams, err)
		return
	}

	if req.Content == "" || len(req.Content) > 2000 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddMessage", domain.ErrRequestParams, nil)
		return
	}

	userID, guestID := h.getUserIdentifiers(ctx)
	userRole := h.getUserRole(ctx)

	err := h.uc.AddMessage(
		ctx,
		ticketID,
		userID,
		guestID,
		userRole,
		req.Content,
	)

	if err != nil {
		log.ErrorContext(ctx, "AddMessage usecase failed", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrAccessDenied):
			h.rs.Error(ctx, w, http.StatusForbidden, "AddMessage", err, nil)
		case errors.Is(err, domain.ErrTicketClosed):
			h.rs.Error(ctx, w, http.StatusConflict, "AddMessage", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "AddMessage", domain.ErrInternalServer, err)
		}
		return
	}

	h.rs.Send(ctx, w, http.StatusNoContent, nil)
}

// AddRatingHandler - POST /tickets/{id}/rating
func (h *SupportHandler) AddRatingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPost {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "AddRating", domain.ErrHTTPMethod, nil)
		return
	}

	ticketID := h.extractTicketID(r)
	if ticketID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddRating", domain.ErrRequestParams, nil)
		return
	}

	var req transport.AddRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "AddRating decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddRating", domain.ErrRequestParams, err)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddRating", domain.ErrRequestParams, nil)
		return
	}

	if len(req.Comment) > 500 {
		h.rs.Error(ctx, w, http.StatusBadRequest, "AddRating", domain.ErrRequestParams, nil)
		return
	}

	userID, guestID := h.getUserIdentifiers(ctx)

	err := h.uc.AddRating(
		ctx,
		ticketID,
		userID,
		guestID,
		req.Rating,
		req.Comment,
	)

	if err != nil {
		log.ErrorContext(ctx, "AddRating usecase failed", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrAccessDenied):
			h.rs.Error(ctx, w, http.StatusForbidden, "AddRating", err, nil)
		case errors.Is(err, domain.ErrTicketNotClosed), errors.Is(err, domain.ErrRatingExists):
			h.rs.Error(ctx, w, http.StatusConflict, "AddRating", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "AddRating", domain.ErrInternalServer, err)
		}
		return
	}

	h.rs.Send(ctx, w, http.StatusNoContent, nil)
}

// GetAllTicketsHandler - GET /admin/tickets
func (h *SupportHandler) GetAllTicketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodGet {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetAllTickets", domain.ErrHTTPMethod, nil)
		return
	}

	// Проверка админских прав
	if !h.isAdmin(ctx) {
		h.rs.Error(ctx, w, http.StatusForbidden, "GetAllTickets", domain.ErrForbidden, nil)
		return
	}

	filter := &domain.TicketFilter{
		Limit:  h.getIntParam(r, "limit", 50),
		Offset: h.getIntParam(r, "offset", 0),
	}

	if status := h.getStringParam(r, "status"); status != nil {
		if h.validateEnum(*status, []string{"open", "in_progress", "closed"}) {
			filter.Status = status
		}
	}

	if category := h.getStringParam(r, "category"); category != nil {
		if h.validateEnum(*category, []string{"bug", "feature", "complaint"}) {
			filter.Category = category
		}
	}

	tickets, err := h.uc.GetAllTickets(ctx, filter)
	if err != nil {
		log.ErrorContext(ctx, "GetAllTickets usecase failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetAllTickets", domain.ErrInternalServer, err)
		return
	}

	response := transport.ToTicketResponses(tickets)
	h.rs.Send(ctx, w, http.StatusOK, response)
}

// UpdateTicketStatusHandler - PUT /admin/tickets/{id}/status
func (h *SupportHandler) UpdateTicketStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPut {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "UpdateTicketStatus", domain.ErrHTTPMethod, nil)
		return
	}

	// Проверка админских прав
	if !h.isAdmin(ctx) {
		h.rs.Error(ctx, w, http.StatusForbidden, "UpdateTicketStatus", domain.ErrForbidden, nil)
		return
	}

	ticketID := h.extractAdminTicketID(r)
	if ticketID == "" {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicketStatus", domain.ErrRequestParams, nil)
		return
	}

	var req transport.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(ctx, "UpdateTicketStatus decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicketStatus", domain.ErrRequestParams, err)
		return
	}

	if !h.validateEnum(req.Status, []string{"open", "in_progress", "closed"}) {
		h.rs.Error(ctx, w, http.StatusBadRequest, "UpdateTicketStatus", domain.ErrRequestParams, nil)
		return
	}

	err := h.uc.UpdateTicketStatus(ctx, ticketID, req.Status)
	if err != nil {
		log.ErrorContext(ctx, "UpdateTicketStatus usecase failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "UpdateTicketStatus", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(ctx, w, http.StatusNoContent, nil)
}

// GetStatisticsHandler - GET /admin/statistics
func (h *SupportHandler) GetStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodGet {
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "GetStatistics", domain.ErrHTTPMethod, nil)
		return
	}

	// Проверка админских прав
	if !h.isAdmin(ctx) {
		h.rs.Error(ctx, w, http.StatusForbidden, "GetStatistics", domain.ErrForbidden, nil)
		return
	}

	stats, err := h.uc.GetStatistics(ctx)
	if err != nil {
		log.ErrorContext(ctx, "GetStatistics usecase failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusInternalServerError, "GetStatistics", domain.ErrInternalServer, err)
		return
	}

	response := transport.ToStatisticsResponse(stats)
	h.rs.Send(ctx, w, http.StatusOK, response)
}

// Дополнительные утилиты для admin handlers
func (h *SupportHandler) isAdmin(ctx context.Context) bool {
	role, _ := middlewares.UserRoleFromContext(ctx)
	return role == "admin"
}

func (h *SupportHandler) extractAdminTicketID(r *http.Request) string {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 4 && parts[0] == "admin" && parts[1] == "tickets" {
		return parts[2]
	}
	return ""
}

func (h *SupportHandler) getStringParam(r *http.Request, key string) *string {
	if value := r.URL.Query().Get(key); value != "" {
		return &value
	}
	return nil
}

func (h *SupportHandler) getIntParam(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
