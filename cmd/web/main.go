package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/RehanAthallahAzhar/shopeezy-catalog/db"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/configs"
	dbGenerated "github.com/RehanAthallahAzhar/shopeezy-catalog/internal/db"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/delivery/http/middlewares"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/delivery/http/routes"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/handlers"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/models"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/pkg/grpc/account"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/pkg/logger"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/pkg/redis"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/repositories"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/services"

	accountpb "github.com/RehanAthallahAzhar/shopeezy-protos/pb/account"
	authpb "github.com/RehanAthallahAzhar/shopeezy-protos/pb/auth"
)

func main() {
	log := logger.NewLogger()

	cfg, err := configs.LoadConfig(log)
	if err != nil {
		log.Fatalf("FATAL: Gagal memuat konfigurasi: %v", err)
	}

	log.Printf(">>>> BUKTI PASSWORD YANG DIBACA: '%s'", cfg.Database.Password)

	dbCredential := models.Credential{
		Host:         cfg.Database.Host,
		Username:     cfg.Database.User,
		Password:     cfg.Database.Password,
		DatabaseName: cfg.Database.Name,
		Port:         cfg.Database.Port,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to DB
	conn, err := db.Connect(ctx, &dbCredential)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	// Migrations
	log.Println("Running database migrations...")

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbCredential.Username,
		dbCredential.Password,
		dbCredential.Host,
		dbCredential.Port,
		dbCredential.DatabaseName,
	)

	m, err := migrate.New(
		"file://../../db/migrations", // Local Path
		// "file://db/migrations", // Container Path
		connectionString,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations ran successfully.")
	defer conn.Close()

	// Init SQLC query
	sqlcQueries := dbGenerated.New(conn)

	// setup redis
	redisClient, err := redis.NewRedisClient()
	if err != nil {
		log.Fatalf("Gagal menginisialisasi klien Redis: %v", err)
	}
	defer redisClient.Close() // Pastikan koneksi Redis ditutup

	// gRPC Account & Auth
	accountConn := createGrpcConnection(cfg.GRPC.AccountServiceAddress, log)
	defer accountConn.Close()

	accountClient := accountpb.NewAccountServiceClient(accountConn)
	authClient := authpb.NewAuthServiceClient(accountConn)
	authClientWrapper := account.NewAuthClientFromService(authClient, accountConn)

	// Publisher Rabbitmq
	rabbitMQURL := cfg.RabbitMQ.URL
	rabbitConn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer rabbitChannel.Close()

	// // Buat instance publisher
	// orderPublisher, err := messaging.NewRabbitMQPublisher(rabbitChannel)
	// if err != nil {
	// 	log.Fatalf("Gagal membuat order event publisher: %v", err)
	// }

	// Depedency Ijection
	productsRepo := repositories.NewProductRepository(sqlcQueries, log)
	cartsRepo := repositories.NewCartRepository(redisClient, log)

	validate := validator.New()
	productService := services.NewProductService(productsRepo, redisClient, validate, log)
	cartService := services.NewCartService(cartsRepo, productService, redisClient, accountClient, log)

	handler := handlers.NewHandler(productService, cartService, log)

	// middleware
	authMiddleware := middlewares.AuthMiddleware(authClientWrapper, log)

	e := echo.New()

	routes.InitRoutes(e, handler, authMiddleware)

	// Middleware default
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	log.Printf("Server Echo for Inventory-Cart is listening on port %s", cfg.Server.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Server.Port))
}

func createGrpcConnection(url string, log *logrus.Logger) *grpc.ClientConn {
	// Gunakan grpc.Dial, yang modern dan non-blocking
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// Gunakan Fatalf di sini karena jika koneksi gagal saat startup,
		// aplikasi tidak bisa berjalan dengan benar.
		log.Fatalf("Failed to connect to the gRPC service at %s: %v", url, err)
	}
	log.Printf("Connect to the gRPC Service at %s", url)
	return conn
}
