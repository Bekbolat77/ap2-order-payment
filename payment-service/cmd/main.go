package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	repository "example.com/payment-service/internal/repository"
	httptransport "example.com/payment-service/internal/transport/http"
	"example.com/payment-service/internal/usecase"
)

func main() {
	port := getEnv("PORT", "8082")
	dbDSN := getEnv("PAYMENT_DB_DSN", "root:root@tcp(localhost:3306)/payment_db?parseTime=true")

	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo)
	handler := httptransport.NewHandler(paymentUsecase)

	router := gin.Default()
	handler.RegisterRoutes(router)

	log.Printf("payment-service running on :%s", port)
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
