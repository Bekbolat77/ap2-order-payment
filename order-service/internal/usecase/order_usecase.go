package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	paymentpb "github.com/Bekbolat77/ap2-payment-generated/paymentpb"

	"example.com/order-service/internal/domain"
)

var (
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrdersNotFound           = errors.New("orders not found")
	ErrInvalidOrderAmount       = errors.New("amount must be greater than 0")
	ErrOnlyPendingCanBeCanceled = errors.New("only pending orders can be cancelled")
	ErrPaidOrderCannotBeCancel  = errors.New("paid orders cannot be cancelled")
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error)
}

type OrderUsecase struct {
	orderRepo     OrderRepository
	paymentClient paymentpb.PaymentServiceClient
}

type CreateOrderInput struct {
	CustomerID string
	ItemName   string
	Amount     int64
}

func NewOrderUsecase(orderRepo OrderRepository, paymentClient paymentpb.PaymentServiceClient) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:     orderRepo,
		paymentClient: paymentClient,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidOrderAmount
	}

	order := &domain.Order{
		ID:         generateID(),
		CustomerID: strings.TrimSpace(input.CustomerID),
		ItemName:   strings.TrimSpace(input.ItemName),
		Amount:     input.Amount,
		Status:     domain.OrderStatusPending,
		CreatedAt:  time.Now().UTC(),
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	paymentResp, err := u.paymentClient.ProcessPayment(ctx, &paymentpb.PaymentRequest{
		OrderId: order.ID,
		Amount:  order.Amount,
	})
	if err != nil {
		return nil, err
	}

	if paymentResp.GetStatus() == "Authorized" {
		order.Status = domain.OrderStatusPaid
	} else {
		order.Status = domain.OrderStatusFailed
	}

	if err := u.orderRepo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, err
	}

	return order, nil
}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (u *OrderUsecase) GetOrdersByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error) {
	orders, err := u.orderRepo.GetByCustomerID(ctx, strings.TrimSpace(customerID))
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}
	return orders, nil
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	if order.Status == domain.OrderStatusPaid {
		return nil, ErrPaidOrderCannotBeCancel
	}
	if order.Status != domain.OrderStatusPending {
		return nil, ErrOnlyPendingCanBeCanceled
	}

	if err := u.orderRepo.UpdateStatus(ctx, order.ID, domain.OrderStatusCancelled); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatusCancelled
	return order, nil
}

func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
