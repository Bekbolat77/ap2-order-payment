package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	paymentpb "github.com/Bekbolat77/ap2-payment-generated/paymentpb"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"

	"example.com/payment-service/internal/broker"
	repository "example.com/payment-service/internal/repository"
	grpctransport "example.com/payment-service/internal/transport/grpc"
	"example.com/payment-service/internal/usecase"
)

func main() {
	grpcPort := getEnv("GRPC_PORT", "50051")
	dbDSN := getEnv("PAYMENT_DB_DSN", "root@tcp(127.0.0.1:3306)/payment_db?parseTime=true")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	queueName := getEnv("PAYMENT_QUEUE", "payment.completed")

	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	publisher, err := broker.NewRabbitMQPublisher(rabbitURL, queueName)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer publisher.Close()

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo, publisher)
	grpcServerHandler := grpctransport.NewServer(paymentUsecase)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(server, grpcServerHandler)

	log.Printf("payment-service gRPC running on :%s", grpcPort)
	log.Printf("payment-service publishes events to queue: %s", queueName)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
