package grpc

import (
	"context"
	"time"

	orderpb "github.com/Bekbolat77/ap2-order-generated/orderpb"

	"example.com/order-service/internal/domain"
)

type OrderReader interface {
	GetOrderByID(ctx context.Context, id string) (*domain.Order, error)
}

type Server struct {
	orderpb.UnimplementedOrderServiceServer
	orderUsecase OrderReader
}

func NewServer(orderUsecase OrderReader) *Server {
	return &Server{
		orderUsecase: orderUsecase,
	}
}

func (s *Server) SubscribeToOrderUpdates(
	req *orderpb.OrderRequest,
	stream orderpb.OrderService_SubscribeToOrderUpdatesServer,
) error {
	ctx := stream.Context()
	orderID := req.GetOrderId()

	var lastStatus string

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		order, err := s.orderUsecase.GetOrderByID(ctx, orderID)
		if err == nil && order != nil {
			if order.Status != lastStatus {
				lastStatus = order.Status

				err = stream.Send(&orderpb.OrderStatusUpdate{
					OrderId:    order.ID,
					Status:     order.Status,
					CustomerId: order.CustomerID,
					ItemName:   order.ItemName,
					Amount:     order.Amount,
					CreatedAt:  order.CreatedAt.Format(time.RFC3339),
				})
				if err != nil {
					return err
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
