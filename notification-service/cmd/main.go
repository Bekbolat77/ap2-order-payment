package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/notification-service/internal/consumer"
)

func main() {
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	queueName := getEnv("PAYMENT_QUEUE", "payment.completed")

	cons, err := consumer.NewRabbitMQConsumer(rabbitURL, queueName)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer cons.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := cons.Start(ctx); err != nil {
			log.Fatalf("consumer error: %v", err)
		}
	}()

	log.Println("notification-service is running...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("shutting down notification-service...")
	cancel()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
