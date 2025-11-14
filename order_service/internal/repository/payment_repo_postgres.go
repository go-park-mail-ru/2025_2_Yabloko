package repository

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/logger"
	"context"
	_ "embed"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:embed sql/payment/insert.sql
var insertPayment string

//go:embed sql/payment/get_by_yookassa_id.sql
var getPaymentByYookassaID string

//go:embed sql/payment/get_by_order_id.sql
var getPaymentByOrderID string

//go:embed sql/payment/update_status.sql
var updatePaymentStatus string

//go:embed sql/payment/get_by_id.sql
var getPaymentByID string

type PaymentRepoPostgres struct {
	db PgxIface
}

func NewPaymentRepoPostgres(db PgxIface) *PaymentRepoPostgres {
	return &PaymentRepoPostgres{
		db: db,
	}
}

func (r *PaymentRepoPostgres) Create(ctx context.Context, payment *domain.Payment) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo Create payment start",
		slog.String("order_id", payment.OrderID),
		slog.String("yookassa_id", payment.YookassaID))

	payment.ID = uuid.New().String()

	_, err := r.db.Exec(ctx, insertPayment,
		payment.ID,
		payment.OrderID,
		payment.YookassaID,
		payment.Status,
		payment.Amount,
		payment.Currency,
		payment.Description,
		payment.Metadata,
	)
	if err != nil {
		log.ErrorContext(ctx, "repo Create payment failed",
			slog.String("order_id", payment.OrderID),
			slog.String("yookassa_id", payment.YookassaID),
			slog.Any("err", err))
		return domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo Create payment success",
		slog.String("payment_id", payment.ID),
		slog.String("order_id", payment.OrderID))
	return nil
}

func (r *PaymentRepoPostgres) GetByYookassaID(ctx context.Context, yookassaID string) (*domain.Payment, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetByYookassaID start", slog.String("yookassa_id", yookassaID))

	var payment domain.Payment
	err := r.db.QueryRow(ctx, getPaymentByYookassaID, yookassaID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.YookassaID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.Description,
		&payment.Metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo GetByYookassaID not found", slog.String("yookassa_id", yookassaID))
			return nil, domain.ErrRowsNotFound
		}
		log.ErrorContext(ctx, "repo GetByYookassaID query failed",
			slog.String("yookassa_id", yookassaID),
			slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo GetByYookassaID success", slog.String("yookassa_id", yookassaID))
	return &payment, nil
}

func (r *PaymentRepoPostgres) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetByOrderID start", slog.String("order_id", orderID))

	var payment domain.Payment
	err := r.db.QueryRow(ctx, getPaymentByOrderID, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.YookassaID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.Description,
		&payment.Metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo GetByOrderID not found", slog.String("order_id", orderID))
			return nil, domain.ErrRowsNotFound
		}
		log.ErrorContext(ctx, "repo GetByOrderID query failed",
			slog.String("order_id", orderID),
			slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo GetByOrderID success", slog.String("order_id", orderID))
	return &payment, nil
}

func (r *PaymentRepoPostgres) UpdateStatus(ctx context.Context, yookassaID string, status domain.PaymentStatus) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo UpdateStatus start",
		slog.String("yookassa_id", yookassaID),
		slog.String("status", string(status)))

	res, err := r.db.Exec(ctx, updatePaymentStatus, yookassaID, status)
	if err != nil {
		log.ErrorContext(ctx, "repo UpdateStatus failed",
			slog.String("yookassa_id", yookassaID),
			slog.String("status", string(status)),
			slog.Any("err", err))
		return domain.ErrInternalServer
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		log.DebugContext(ctx, "repo UpdateStatus: no rows updated (possible race)",
			slog.String("yookassa_id", yookassaID),
			slog.String("status", string(status)))
		// Не ошибка — просто статус уже такой
		return nil
	}

	log.DebugContext(ctx, "repo UpdateStatus success",
		slog.String("yookassa_id", yookassaID),
		slog.String("status", string(status)),
		slog.Int64("rows_affected", rowsAffected))
	return nil
}

func (r *PaymentRepoPostgres) GetByID(ctx context.Context, paymentID string) (*domain.Payment, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetByID start", slog.String("payment_id", paymentID))

	var payment domain.Payment
	err := r.db.QueryRow(ctx, getPaymentByID, paymentID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.YookassaID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.Description,
		&payment.Metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo GetByID not found", slog.String("payment_id", paymentID))
			return nil, domain.ErrRowsNotFound
		}
		log.ErrorContext(ctx, "repo GetByID query failed",
			slog.String("payment_id", paymentID),
			slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo GetByID success", slog.String("payment_id", paymentID))
	return &payment, nil
}
