package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	// Sesuaikan import path ini dengan module name Anda
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/databases" // Ini mungkin perlu disesuaikan jika nama repo berubah
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/handlers"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/authclient" // Impor authclient
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/redisclient"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/repositories"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/routes"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	portStr := os.Getenv("DB_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("Invalid DB_PORT in .env file or not set: " + err.Error())
	}

	dbCredential := models.Credential{
		Host:         os.Getenv("DB_HOST"),
		Username:     os.Getenv("DB_USER"),
		Password:     os.Getenv("DB_PASSWORD"),
		DatabaseName: os.Getenv("DB_NAME"),
		Port:         port,
	}

	dbInstance := databases.NewDB()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := dbInstance.Connect(ctx, &dbCredential)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer func() {
		sqlDB, err := conn.DB()
		if err != nil {
			log.Printf("Error getting underlying DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	err = conn.AutoMigrate(&models.Product{}, &models.Cart{}, &models.Order{}, &models.OrderItem{}) // Pastikan Order dan OrderItem juga dimigrasi
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// --- Inisialisasi Klien Redis ---
	redisClient, err := redisclient.NewRedisClient()
	if err != nil {
		log.Fatalf("Gagal menginisialisasi klien Redis: %v", err)
	}
	defer redisClient.Close() // Pastikan koneksi Redis ditutup

	// --- Inisialisasi gRPC Auth Client ---
	accountGRPCServerAddress := os.Getenv("ACCOUNT_GRPC_SERVER_ADDRESS")
	if accountGRPCServerAddress == "" {
		accountGRPCServerAddress = "localhost:50051"
	}

	authClient, err := authclient.NewAuthClient(accountGRPCServerAddress)
	if err != nil {
		log.Fatalf("Gagal membuat klien gRPC Auth: %v", err)
	}
	defer authClient.Close()

	// repo
	productsRepo := repositories.NewProductRepository(conn, redisClient)
	cartsRepo := repositories.NewCartRepository(conn, productsRepo, redisClient)

	// service
	validate := validator.New()

	productService := services.NewProductService(productsRepo, validate)
	cartService := services.NewCartService(cartsRepo, productsRepo)

	// --- Inisialisasi Echo Server dan Handler ---
	e := echo.New()

	// Middleware default
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	handler := handlers.NewHandler(authClient, productsRepo, cartsRepo, productService, cartService)
	routes.InitRoutes(e, handler)

	log.Printf("Server Echo Cashier App mendengarkan di port 1323")
	e.Logger.Fatal(e.Start(":1323"))
}
