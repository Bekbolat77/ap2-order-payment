package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"example.com/order-service/internal/domain"
	"example.com/order-service/internal/usecase"
)

type Handler struct {
	orderUsecase *usecase.OrderUsecase
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
}

type orderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func NewHandler(orderUsecase *usecase.OrderUsecase) *Handler {
	return &Handler{
		orderUsecase: orderUsecase,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/orders", h.CreateOrder)
	router.GET("/orders/:id", h.GetOrderByID)

	router.GET("/orders/customer/:customer_id", h.GetOrdersByCustomerID)

	router.PATCH("/orders/:id/cancel", h.CancelOrder)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	order, err := h.orderUsecase.CreateOrder(c.Request.Context(), usecase.CreateOrderInput{
		CustomerID: req.CustomerID,
		ItemName:   req.ItemName,
		Amount:     req.Amount,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidAmount):
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case errors.Is(err, usecase.ErrPaymentUnavailable):
			c.JSON(nethttp.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(nethttp.StatusCreated, toOrderResponse(order))
}

func (h *Handler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.orderUsecase.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusOK, toOrderResponse(order))
}


func (h *Handler) GetOrdersByCustomerID(c *gin.Context) {
	customerID := c.Param("customer_id")

	orders, err := h.orderUsecase.GetOrdersByCustomerID(c.Request.Context(), customerID)
	if err != nil {
		if errors.Is(err, usecase.ErrOrdersNotFound) {
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	
	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		orderCopy := order
		response = append(response, toOrderResponse(&orderCopy))
	}

	c.JSON(nethttp.StatusOK, response)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.orderUsecase.CancelOrder(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrOrderNotFound):
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
			return
		case errors.Is(err, usecase.ErrPaidOrderCannotBeCancel),
			errors.Is(err, usecase.ErrOnlyPendingCanBeCanceled):
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(nethttp.StatusOK, toOrderResponse(order))
}

func toOrderResponse(order *domain.Order) orderResponse {
	return orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
