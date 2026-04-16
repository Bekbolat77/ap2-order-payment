package repository

import (
	"context"
	"database/sql"
	"errors"

	"example.com/payment-service/internal/domain"
	"example.com/payment-service/internal/usecase"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
	)
	return err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments
		WHERE order_id = ?
		LIMIT 1
	`

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, usecase.ErrPaymentNotFound
		}
		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) ListByStatus(ctx context.Context, status string) ([]domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments
		WHERE status = ?
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]domain.Payment, 0)
	for rows.Next() {
		var payment domain.Payment
		if err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.TransactionID,
			&payment.Amount,
			&payment.Status,
		); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}
