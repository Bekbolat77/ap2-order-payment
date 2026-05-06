package domain

type NotificationEvent struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	EventID       string  `json:"event_id"`
}
