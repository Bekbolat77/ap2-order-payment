package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"example.com/payment-service/internal/domain"
	"example.com/payment-service/internal/usecase"
)

type Handler struct {
	paymentUsecase *usecase.PaymentUsecase
}

type createPaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func NewHandler(paymentUsecase *usecase.PaymentUsecase) *Handler {
	return &Handler{
		paymentUsecase: paymentUsecase,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/payments", h.CreatePayment)
	router.GET("/payments/:order_id", h.GetPaymentByOrderID)
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	payment, err := h.paymentUsecase.CreatePayment(c.Request.Context(), usecase.CreatePaymentInput{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidAmount):
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(nethttp.StatusCreated, toPaymentResponse(payment))
}

func (h *Handler) GetPaymentByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.paymentUsecase.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, usecase.ErrPaymentNotFound) {
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusOK, toPaymentResponse(payment))
}

func toPaymentResponse(payment *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
	}
}
