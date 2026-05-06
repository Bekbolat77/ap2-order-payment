package grpc

import (
	"context"

	paymentpb "github.com/Bekbolat77/ap2-payment-generated/paymentpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"example.com/payment-service/internal/domain"
	"example.com/payment-service/internal/usecase"
)

type Server struct {
	paymentpb.UnimplementedPaymentServiceServer
	paymentUsecase *usecase.PaymentUsecase
}

func NewServer(paymentUsecase *usecase.PaymentUsecase) *Server {
	return &Server{
		paymentUsecase: paymentUsecase,
	}
}
func (s *Server) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	payment, err := s.paymentUsecase.CreatePayment(ctx, usecase.CreatePaymentInput{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	})
	if err != nil {
		switch err {
		case usecase.ErrInvalidAmount:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toPaymentResponse(payment), nil
}

func (s *Server) ListPayments(ctx context.Context, req *paymentpb.ListPaymentsRequest) (*paymentpb.ListPaymentsResponse, error) {
	payments, err := s.paymentUsecase.ListPaymentsByStatus(ctx, req.GetStatus())
	if err != nil {
		switch err {
		case usecase.ErrInvalidStatus:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := &paymentpb.ListPaymentsResponse{
		Payments: make([]*paymentpb.PaymentResponse, 0, len(payments)),
	}

	for _, p := range payments {
		paymentCopy := p
		resp.Payments = append(resp.Payments, toPaymentResponse(&paymentCopy))
	}

	return resp, nil
}

func toPaymentResponse(payment *domain.Payment) *paymentpb.PaymentResponse {
	return &paymentpb.PaymentResponse{
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
	}
}
