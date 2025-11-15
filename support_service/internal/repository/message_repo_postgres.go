package repository

import (
	"context"
	"log/slog"

	"support_service/internal/domain"
	"support_service/pkg/logger"
)

type MessageRepoPostgres struct {
	db PgxIface
}

func NewMessageRepoPostgres(db PgxIface) *MessageRepoPostgres {
	return &MessageRepoPostgres{db: db}
}

func (r *MessageRepoPostgres) CreateMessage(ctx context.Context, m *domain.Message) error {
	log := logger.FromContext(ctx)

	// ✅ Объект УЖЕ готовый из UseCase (с ID и временем)
	query := `INSERT INTO support_message (id, ticket_id, user_id, guest_id, user_role, content, created_at)
			  VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.Exec(ctx, query, m.ID, m.TicketID, m.UserID, m.GuestID, m.UserRole, m.Content, m.CreatedAt)
	if err != nil {
		log.ErrorContext(ctx, "CreateMessage db error", slog.Any("err", err))
		return err
	}
	return nil
}

func (r *MessageRepoPostgres) GetMessagesByTicket(ctx context.Context, ticketID string) ([]*domain.Message, error) {
	log := logger.FromContext(ctx)
	query := `SELECT id, ticket_id, user_id, guest_id, user_role, content, created_at FROM support_message WHERE ticket_id=$1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		log.ErrorContext(ctx, "GetMessagesByTicket db error", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.TicketID, &m.UserID, &m.GuestID, &m.UserRole, &m.Content, &m.CreatedAt); err != nil {
			log.ErrorContext(ctx, "GetMessagesByTicket scan error", slog.Any("err", err))
			return nil, err
		}
		messages = append(messages, &m)
	}
	return messages, nil
}
