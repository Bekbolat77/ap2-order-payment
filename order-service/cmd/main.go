package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	repository "example.com/order-service/internal/repository"
	httptransport "example.com/order-service/internal/transport/http"
	"example.com/order-service/internal/usecase"
)

func main() {
	port := getEnv("PORT", "8081")
	dbDSN := getEnv("ORDER_DB_DSN", "root:root@tcp(localhost:3306)/order_db?parseTime=true")
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082")

	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	orderRepo := repository.NewOrderRepository(db)

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	orderUsecase := usecase.NewOrderUsecase(orderRepo, httpClient, paymentServiceURL)

	handler := httptransport.NewHandler(orderUsecase)

	router := gin.Default()
	handler.RegisterRoutes(router)

	log.Printf("order-service running on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
