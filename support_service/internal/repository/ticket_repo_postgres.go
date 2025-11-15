package repository

import (
	"context"
	"errors"
	"log/slog"

	"apple_backend/pkg/logger"
	"apple_backend/support_service/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type TicketRepoPostgres struct {
	db PgxIface
}

func NewTicketRepoPostgres(db PgxIface) *TicketRepoPostgres {
	return &TicketRepoPostgres{db: db}
}

func (r *TicketRepoPostgres) CreateTicket(ctx context.Context, t *domain.Ticket) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "CreateTicket start")

	query := `
	INSERT INTO support_ticket
	(id, user_id, guest_id, user_name, user_email, category, status, priority, title, description, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`
	_, err := r.db.Exec(ctx, query,
		t.ID, t.UserID, t.GuestID, t.UserName, t.UserEmail,
		t.Category, t.Status, t.Priority, t.Title, t.Description,
		t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.WarnContext(ctx, "CreateTicket duplicate", slog.Any("err", err))
			return domain.ErrRequestParams
		}
		log.ErrorContext(ctx, "CreateTicket db error", slog.Any("err", err))
		return err
	}

	log.DebugContext(ctx, "CreateTicket success", slog.String("id", t.ID))
	return nil
}

func (r *TicketRepoPostgres) GetTicket(ctx context.Context, id string) (*domain.Ticket, error) {
	log := logger.FromContext(ctx)
	query := `SELECT id, user_id, guest_id, user_name, user_email, category, status, priority, title, description, created_at, updated_at
			  FROM support_ticket WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var t domain.Ticket
	err := row.Scan(
		&t.ID, &t.UserID, &t.GuestID, &t.UserName, &t.UserEmail,
		&t.Category, &t.Status, &t.Priority, &t.Title, &t.Description,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		log.ErrorContext(ctx, "GetTicket db error", slog.Any("err", err))
		return nil, domain.ErrRowsNotFound
	}
	return &t, nil
}

func (r *TicketRepoPostgres) GetUserTickets(ctx context.Context, userID, guestID *string) ([]*domain.Ticket, error) {
	log := logger.FromContext(ctx)
	query := `SELECT id, user_id, guest_id, user_name, user_email, category, status, priority, title, description, created_at, updated_at
			  FROM support_ticket WHERE (user_id=$1 OR guest_id=$2) ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID, guestID)
	if err != nil {
		log.ErrorContext(ctx, "GetUserTickets db error", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.GuestID, &t.UserName, &t.UserEmail,
			&t.Category, &t.Status, &t.Priority, &t.Title, &t.Description,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			log.ErrorContext(ctx, "GetUserTickets scan error", slog.Any("err", err))
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *TicketRepoPostgres) UpdateTicketStatus(ctx context.Context, id, status string) error {
	log := logger.FromContext(ctx)
	query := `UPDATE support_ticket SET status=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		log.ErrorContext(ctx, "UpdateTicketStatus db error", slog.Any("err", err))
		return err
	}
	return nil
}

func (r *TicketRepoPostgres) GetUserByID(ctx context.Context, userID string) (*domain.UserInfo, error) {
	log := logger.FromContext(ctx)
	query := `SELECT id, name, email FROM users WHERE id=$1`
	row := r.db.QueryRow(ctx, query, userID)

	var u domain.UserInfo
	err := row.Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		log.ErrorContext(ctx, "GetUserByID db error", slog.Any("err", err))
		return nil, domain.ErrRowsNotFound
	}
	return &u, nil
}

func (r *TicketRepoPostgres) GetAllTickets(ctx context.Context, filter *domain.TicketFilter) ([]*domain.Ticket, error) {
	log := logger.FromContext(ctx)

	query := `
    SELECT id, user_id, guest_id, user_name, user_email, category, status, priority, 
           title, description, created_at, updated_at
    FROM support_ticket 
    WHERE 1=1
    `
	var args []interface{}

	if filter.Status != nil {
		query += " AND status = $1"
		args = append(args, *filter.Status)
	}
	if filter.Category != nil {
		query += " AND category = $2"
		args = append(args, *filter.Category)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT $3"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET $4"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.ErrorContext(ctx, "GetAllTickets db error", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.GuestID, &t.UserName, &t.UserEmail,
			&t.Category, &t.Status, &t.Priority, &t.Title, &t.Description,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *TicketRepoPostgres) GetStatistics(ctx context.Context) (*domain.Statistics, error) {

	stats := &domain.Statistics{
		TicketsByCategory: make(map[string]int),
	}

	// Total tickets
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM support_ticket").Scan(&stats.TotalTickets)
	if err != nil {
		return nil, err
	}

	// Tickets by status
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM support_ticket WHERE status = 'open'").Scan(&stats.OpenTickets)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM support_ticket WHERE status = 'in_progress'").Scan(&stats.InProgressTickets)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM support_ticket WHERE status = 'closed'").Scan(&stats.ClosedTickets)

	// Tickets by category
	rows, err := r.db.Query(ctx, "SELECT category, COUNT(*) FROM support_ticket GROUP BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		stats.TicketsByCategory[category] = count
	}

	return stats, nil
}

func (r *TicketRepoPostgres) UpdateTicket(ctx context.Context, t *domain.Ticket) error {
	log := logger.FromContext(ctx)

	query := `
    UPDATE support_ticket 
    SET title=$1, description=$2, category=$3, updated_at=$4
    WHERE id=$5
    `
	_, err := r.db.Exec(ctx, query,
		t.Title, t.Description, t.Category, t.UpdatedAt, t.ID)
	if err != nil {
		log.ErrorContext(ctx, "UpdateTicket db error", slog.Any("err", err))
		return err
	}
	return nil
}
