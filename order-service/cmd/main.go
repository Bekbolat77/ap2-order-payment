package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	orderpb "github.com/Bekbolat77/ap2-order-generated/orderpb"
	paymentpb "github.com/Bekbolat77/ap2-payment-generated/paymentpb"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	repository "example.com/order-service/internal/repository"
	grpctransport "example.com/order-service/internal/transport/grpc"
	httptransport "example.com/order-service/internal/transport/http"
	"example.com/order-service/internal/usecase"
)

func main() {
	port := getEnv("PORT", "8081")
	grpcPort := getEnv("ORDER_GRPC_PORT", "50052")
	dbDSN := getEnv("ORDER_DB_DSN", "root@tcp(127.0.0.1:3306)/order_db?parseTime=true")
	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:50051")

	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	conn, err := grpc.Dial(paymentGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to payment-service gRPC: %v", err)
	}
	defer conn.Close()

	paymentClient := paymentpb.NewPaymentServiceClient(conn)

	orderRepo := repository.NewOrderRepository(db)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, paymentClient)

	// REST server
	handler := httptransport.NewHandler(orderUsecase)
	router := gin.Default()
	handler.RegisterRoutes(router)

	// gRPC streaming server
	grpcHandler := grpctransport.NewServer(orderUsecase)
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, grpcHandler)

	go func() {
		log.Printf("order-service gRPC streaming running on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve order gRPC: %v", err)
		}
	}()

	log.Printf("order-service REST running on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
