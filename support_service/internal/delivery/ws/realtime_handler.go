package ws

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/support_service/internal/domain"
	"apple_backend/support_service/internal/usecase"

	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SupportUsecaseInterface interface {
	GetTicket(ctx context.Context, id string) (*domain.Ticket, error)
}

// События WebSocket
type WSMessage struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type TicketUpdateEvent struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageCreatedEvent struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	UserID    *string   `json:"user_id,omitempty"`
	GuestID   *string   `json:"guest_id,omitempty"`
	UserRole  string    `json:"user_role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Broadcaster — in-memory шина (заменить на Redis когда нить)
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[string]map[*wsClient]bool // ticketID → []*wsClient
}

type wsClient struct {
	conn     *websocket.Conn
	send     chan *WSMessage
	ticketID string
	isAdmin  bool
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]map[*wsClient]bool),
	}
}

func (b *Broadcaster) Subscribe(ticketID string, client *wsClient) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[ticketID]; !ok {
		b.clients[ticketID] = make(map[*wsClient]bool)
	}
	b.clients[ticketID][client] = true
}

func (b *Broadcaster) Unsubscribe(ticketID string, client *wsClient) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if clients, ok := b.clients[ticketID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(b.clients, ticketID)
		}
	}
}

func (b *Broadcaster) Publish(ticketID string, msg *WSMessage) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if clients, ok := b.clients[ticketID]; ok {
		for client := range clients {
			select {
			case client.send <- msg:
			default:
				close(client.send)
			}
		}
	}
}

// RealtimeHandler
type RealtimeHandler struct {
	uc          SupportUsecaseInterface
	broadcaster *Broadcaster
	upgrader    websocket.Upgrader
	rs          *http_response.ResponseSender
}

func NewRealtimeHandler(uc SupportUsecaseInterface) *RealtimeHandler {
	return &RealtimeHandler{
		uc:          uc,
		broadcaster: NewBroadcaster(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: настроить в продакшене
				return true
			},
		},
		rs: http_response.NewResponseSender(logger.Global()),
	}
}

// NewRealtimeRouter — регистрирует роуты (как в store_service)
func NewRealtimeRouter(mux *http.ServeMux, uc usecase.TicketRepository) {
	// Оборачиваем репо в юзкейс (минимальный — только GetTicket)
	minimalUC := &minimalUsecase{repo: uc}
	handler := NewRealtimeHandler(minimalUC)
	mux.HandleFunc("/ws/ticket/", handler.SubscribeTicket)
}

// minimalUsecase — адаптер для WS (только GetTicket)
type minimalUsecase struct {
	repo usecase.TicketRepository
}

func (m *minimalUsecase) GetTicket(ctx context.Context, id string) (*domain.Ticket, error) {
	return m.repo.GetTicket(ctx, id)
}

// SubscribeTicket — обрабатывает /ws/ticket/{ticketID}
//
// TODO: требуется мидлвара, которая:
//   - парсит JWT → кладёт в context: userID (string), isAdmin (bool)
//   - читает X-Guest-ID → кладёт в context: guestID (string)
//   - передаётся через context.WithValue или кастомный middleware chain
func (h *RealtimeHandler) SubscribeTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler SubscribeTicket start")

	if r.Method != http.MethodGet {
		log.WarnContext(ctx, "handler SubscribeTicket wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "SubscribeTicket", domain.ErrHTTPMethod, nil)
		return
	}

	// Извлекаем ticketID из пути: /ws/ticket/abc => "abc"
	ticketID := r.URL.Path[len("/ws/ticket/"):]
	if ticketID == "" {
		log.WarnContext(ctx, "handler SubscribeTicket empty ticketID")
		h.rs.Error(ctx, w, http.StatusBadRequest, "SubscribeTicket", domain.ErrRequestParams, errors.New("ticketID required"))
		return
	}

	// TODO: эти значения ДОЛЖНЫ приходить из мидлвары через context!
	var (
		userID  *string
		guestID *string
		isAdmin bool
	)

	if uid, ok := ctx.Value("userID").(string); ok && uid != "" {
		userID = &uid
	}
	if gid, ok := ctx.Value("guestID").(string); ok && gid != "" {
		guestID = &gid
	}
	if admin, ok := ctx.Value("isAdmin").(bool); ok {
		isAdmin = admin
	}

	// Валидация: нужен хотя бы один идентификатор
	if userID == nil && guestID == nil {
		log.WarnContext(ctx, "handler SubscribeTicket no auth")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "SubscribeTicket", domain.ErrUnauthorized, nil)
		return
	}

	// Проверка доступа к тикету
	_, err := h.validateTicketAccess(ctx, ticketID, userID, guestID)
	if err != nil {
		log.ErrorContext(ctx, "handler SubscribeTicket access validation failed", slog.Any("err", err))
		if errors.Is(err, domain.ErrAccessDenied) {
			h.rs.Error(ctx, w, http.StatusForbidden, "SubscribeTicket", domain.ErrAccessDenied, nil)
			return
		}
		if errors.Is(err, domain.ErrRowsNotFound) {
			h.rs.Error(ctx, w, http.StatusNotFound, "SubscribeTicket", domain.ErrRowsNotFound, nil)
			return
		}
		h.rs.Error(ctx, w, http.StatusInternalServerError, "SubscribeTicket", domain.ErrInternalServer, err)
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.ErrorContext(ctx, "handler SubscribeTicket upgrade failed", slog.Any("err", err))
		return // Upgrade уже записал ответ
	}

	client := &wsClient{
		conn:     conn,
		send:     make(chan *WSMessage, 256),
		ticketID: ticketID,
		isAdmin:  isAdmin,
	}

	h.broadcaster.Subscribe(ticketID, client)
	log.InfoContext(ctx, "handler SubscribeTicket success", slog.String("ticket_id", ticketID))

	// Горутины
	go h.readPump(client, log)
	go h.writePump(client, log)
}

func (h *RealtimeHandler) validateTicketAccess(
	ctx context.Context,
	ticketID string,
	userID, guestID *string,
) (*domain.Ticket, error) {
	ticket, err := h.uc.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	isOwner := false
	if userID != nil && ticket.UserID != nil && *userID == *ticket.UserID {
		isOwner = true
	}
	if guestID != nil && ticket.GuestID != nil && *guestID == *ticket.GuestID {
		isOwner = true
	}

	if !isOwner {
		return nil, domain.ErrAccessDenied
	}
	return ticket, nil
}

func (h *RealtimeHandler) readPump(client *wsClient, log *slog.Logger) {
	defer func() {
		h.broadcaster.Unsubscribe(client.ticketID, client)
		client.conn.Close()
		close(client.send)
		log.Info("WS client disconnected")
	}()

	client.conn.SetReadLimit(1024)
	_ = client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		_ = client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debug("WS close", "err", err)
			} else {
				log.Warn("WS read error", "err", err)
			}
			break
		}
	}
}

func (h *RealtimeHandler) writePump(client *wsClient, log *slog.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			// 🔑 Фильтрация: ticket.updated — только админу
			if msg.Event == "ticket.updated" && !client.isAdmin {
				continue
			}

			_ = client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if err := json.NewEncoder(w).Encode(msg); err != nil {
				log.Warn("WS encode failed", "err", err)
				return
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// =============
// Публичные методы для вызова из других хендлеров (HTTP)

// OnTicketUpdated — вызывать после успешного UpdateTicketStatus
// Пример использования в HTTP-хендлере:
//
//	err := uc.UpdateTicketStatus(ctx, ticketID, status)
//	if err == nil {
//	    ticket, _ := uc.GetTicket(ctx, ticketID)
//	    realtimeHandler.OnTicketUpdated(ticketID, status, ticket.Priority)
//	}
func (h *RealtimeHandler) OnTicketUpdated(ticketID, status, priority string) {
	msg := &WSMessage{
		Event: "ticket.updated",
		Payload: TicketUpdateEvent{
			ID:        ticketID,
			Status:    status,
			Priority:  priority,
			UpdatedAt: time.Now().UTC(),
		},
	}
	h.broadcaster.Publish(ticketID, msg)
}

// OnMessageCreated — вызывать после успешного AddMessage
// Пример:
//
//	err := uc.AddMessage(ctx, ticketID, userID, guestID, userRole, content)
//	if err == nil {
//	    realtimeHandler.OnMessageCreated(&MessageCreatedEvent{
//	        ID:        uuid.NewString(),
//	        TicketID:  ticketID,
//	        UserID:    userID,
//	        GuestID:   guestID,
//	        UserRole:  userRole,
//	        Content:   content,
//	        CreatedAt: time.Now().UTC(),
//	    })
//	}
func (h *RealtimeHandler) OnMessageCreated(event *MessageCreatedEvent) {
	msg := &WSMessage{
		Event:   "message.created",
		Payload: event,
	}
	h.broadcaster.Publish(event.TicketID, msg)
}
