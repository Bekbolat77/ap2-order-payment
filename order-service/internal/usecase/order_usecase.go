package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"example.com/order-service/internal/domain"
)

var (
	ErrInvalidAmount            = errors.New("amount must be greater than 0")
	ErrOrderNotFound            = errors.New("order not found")

	ErrOrdersNotFound           = errors.New("orders not found")

	ErrOnlyPendingCanBeCanceled = errors.New("only pending orders can be cancelled")
	ErrPaidOrderCannotBeCancel  = errors.New("paid orders cannot be cancelled")
	ErrPaymentUnavailable       = errors.New("payment service unavailable")
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)

	GetByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error)

	UpdateStatus(ctx context.Context, id string, status string) error
}

type OrderUsecase struct {
	repo              OrderRepository
	httpClient        *http.Client
	paymentServiceURL string
}

type CreateOrderInput struct {
	CustomerID string
	ItemName   string
	Amount     int64
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

func NewOrderUsecase(
	repo OrderRepository,
	httpClient *http.Client,
	paymentServiceURL string,
) *OrderUsecase {
	return &OrderUsecase{
		repo:              repo,
		httpClient:        httpClient,
		paymentServiceURL: strings.TrimRight(paymentServiceURL, "/"),
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	order := &domain.Order{
		ID:         newID(),
		CustomerID: strings.TrimSpace(input.CustomerID),
		ItemName:   strings.TrimSpace(input.ItemName),
		Amount:     input.Amount,
		Status:     domain.OrderStatusPending,
		CreatedAt:  time.Now().UTC(),
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	paymentStatus, err := u.authorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		_ = u.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed)
		order.Status = domain.OrderStatusFailed
		return nil, err
	}

	switch paymentStatus {
	case "Authorized":
		if err := u.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusPaid); err != nil {
			return nil, err
		}
		order.Status = domain.OrderStatusPaid
	case "Declined":
		if err := u.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed); err != nil {
			return nil, err
		}
		order.Status = domain.OrderStatusFailed
	default:
		if err := u.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed); err != nil {
			return nil, err
		}
		order.Status = domain.OrderStatusFailed
	}

	return order, nil
}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	order, err := u.repo.GetByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}



func (u *OrderUsecase) GetOrdersByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error) {
	orders, err := u.repo.GetByCustomerID(ctx, strings.TrimSpace(customerID))
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}
	return orders, nil
}














func (u *OrderUsecase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := u.repo.GetByID(ctx, strings.TrimSpace(id))
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

	if err := u.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusCancelled); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatusCancelled
	return order, nil
}

func (u *OrderUsecase) authorizePayment(ctx context.Context, orderID string, amount int64) (string, error) {
	payload := paymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		u.paymentServiceURL+"/payments",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", ErrPaymentUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return "", ErrPaymentUnavailable
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var pr paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", err
	}

	return pr.Status, nil
}

func newID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
