package broker

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	EventID       string  `json:"event_id"`
}

type RabbitMQPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQPublisher(url string, queueName string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQPublisher{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

func (p *RabbitMQPublisher) PublishPaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
