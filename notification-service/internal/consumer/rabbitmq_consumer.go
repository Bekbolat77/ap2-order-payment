package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"example.com/notification-service/internal/domain"
)

type RabbitMQConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string

	processed map[string]bool
	mu        sync.Mutex
}

func NewRabbitMQConsumer(url string, queueName string) (*RabbitMQConsumer, error) {
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
		true, // durable
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

	return &RabbitMQConsumer{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
		processed: make(map[string]bool),
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"",
		false, // autoAck OFF
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("notification-service started consuming queue:", c.queueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("notification-service stopping consumer...")
			return nil

		case msg := <-msgs:
			var event domain.NotificationEvent

			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("invalid message:", err)
				msg.Nack(false, false)
				continue
			}

			if c.isDuplicate(event.EventID) {
				log.Println("duplicate event ignored:", event.EventID)
				msg.Ack(false)
				continue
			}

			log.Printf(
				"[Notification] Sent email to %s for Order #%s. Amount: %.2f. Status: %s\n",
				event.CustomerEmail,
				event.OrderID,
				event.Amount,
				event.Status,
			)

			c.markProcessed(event.EventID)
			msg.Ack(false)
		}
	}
}

func (c *RabbitMQConsumer) isDuplicate(eventID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.processed[eventID]
}

func (c *RabbitMQConsumer) markProcessed(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.processed[eventID] = true
}

func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
