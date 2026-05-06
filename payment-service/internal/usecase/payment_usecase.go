package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"example.com/payment-service/internal/broker"
	"example.com/payment-service/internal/domain"
)

var (
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidStatus   = errors.New("status must be Authorized or Declined")
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	ListByStatus(ctx context.Context, status string) ([]domain.Payment, error)
}

type PaymentPublisher interface {
	PublishPaymentCompleted(ctx context.Context, event broker.PaymentCompletedEvent) error
}

type PaymentUsecase struct {
	repo      PaymentRepository
	publisher PaymentPublisher
}

func NewPaymentUsecase(repo PaymentRepository, publisher PaymentPublisher) *PaymentUsecase {
	return &PaymentUsecase{
		repo:      repo,
		publisher: publisher,
	}
}

type CreatePaymentInput struct {
	OrderID string
	Amount  int64
}

func (u *PaymentUsecase) CreatePayment(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	status := domain.PaymentStatusAuthorized
	if input.Amount > 100000 {
		status = domain.PaymentStatusDeclined
	}

	payment := &domain.Payment{
		ID:            newID(),
		OrderID:       strings.TrimSpace(input.OrderID),
		TransactionID: newID(),
		Amount:        input.Amount,
		Status:        status,
	}

	if err := u.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	if payment.Status == domain.PaymentStatusAuthorized && u.publisher != nil {
		event := broker.PaymentCompletedEvent{
			EventID:       payment.ID,
			OrderID:       payment.OrderID,
			Amount:        float64(payment.Amount) / 100,
			CustomerEmail: "user@example.com",
			Status:        payment.Status,
		}

		if err := u.publisher.PublishPaymentCompleted(ctx, event); err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func (u *PaymentUsecase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment, err := u.repo.GetByOrderID(ctx, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

func (u *PaymentUsecase) ListPaymentsByStatus(ctx context.Context, status string) ([]domain.Payment, error) {
	status = strings.TrimSpace(status)

	if status != domain.PaymentStatusAuthorized && status != domain.PaymentStatusDeclined {
		return nil, ErrInvalidStatus
	}

	return u.repo.ListByStatus(ctx, status)
}

func newID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
