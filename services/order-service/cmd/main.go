package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jmoiron/sqlx"

	"github.com/nusantara-supermart/pkg/config"
	"github.com/nusantara-supermart/pkg/database"
	"github.com/nusantara-supermart/pkg/messaging"
	"github.com/nusantara-supermart/pkg/middleware"
	"github.com/nusantara-supermart/pkg/response"
	"github.com/nusantara-supermart/services/order-service/internal/logistics"
	"github.com/nusantara-supermart/services/order-service/internal/order"
	"github.com/nusantara-supermart/services/order-service/internal/payment"
)

func connectDatabase(name string, cfg config.DatabaseConfig) *sqlx.DB {
	db, err := database.ConnectMySQL(cfg)
	if err != nil {
		log.Printf("WARNING: %s MySQL connection failed: %v", name, err)
		log.Printf("%s database will be retried when serving requests or initializing", name)
		return nil
	}
	log.Printf("Connected to %s MySQL database successfully!", name)
	return db
}

func closeDatabase(db *sqlx.DB) {
	if db != nil {
		_ = db.Close()
	}
}

func main() {
	cfg := config.LoadConfig()

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("ORDER_SERVICE_PORT")
	}
	if port == "" {
		port = "3004"
	}

	catalogGRPCAddr := os.Getenv("CATALOG_GRPC_ADDR")
	if catalogGRPCAddr == "" {
		if os.Getenv("APP_ENV") == "production" || os.Getenv("DOCKER_CONTAINER") != "" {
			catalogGRPCAddr = "catalog-service:50052"
		} else {
			catalogGRPCAddr = "localhost:50052"
		}
	}

	log.Printf("🚀 Starting Order Service on port %s in [%s] mode...", port, cfg.AppEnv)

	orderDB := connectDatabase("order", cfg.OrderDB)
	defer closeDatabase(orderDB)

	// Inisialisasi gRPC Client ke Catalog Service
	catalogGRPCClient := order.NewCatalogClient(catalogGRPCAddr)
	if catalogGRPCClient != nil {
		defer catalogGRPCClient.Close()
		log.Printf("Order Service connected to Catalog gRPC at %s", catalogGRPCAddr)
	}

	// Inisialisasi RabbitMQ Producer
	rabbitMQ, err := messaging.ConnectRabbitMQ()
	if err != nil {
		log.Printf("WARNING: Gagal koneksi ke RabbitMQ: %v. Berjalan dalam mode fallback.", err)
	} else {
		defer rabbitMQ.Close()
		log.Println("RabbitMQ connected successfully in Order Service")
	}

	// Repositories
	orderRepo := order.NewRepository(orderDB)
	paymentRepo := payment.NewRepository(orderDB)
	logisticsRepo := logistics.NewRepository(orderDB)

	// Services
	orderSvc := order.NewService(orderRepo, catalogGRPCClient)
	if rabbitMQ != nil {
		orderSvc.SetRabbitMQ(rabbitMQ)
	}
	paymentSvc := payment.NewService(paymentRepo, orderSvc)
	logisticsSvc := logistics.NewService(logisticsRepo, orderSvc)

	// Handlers
	orderHandler := order.NewHandler(orderSvc)
	paymentHandler := payment.NewHandler(paymentSvc)
	logisticsHandler := logistics.NewHandler(logisticsSvc)

	// Fiber App
	app := fiber.New(fiber.Config{
		AppName:      "PT Nusantara SuperMart - Order Service v1.0",
		ServerHeader: "NusantaraSuperMart-OrderService",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.SetupCORS(cfg))

	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "PT Nusantara SuperMart Order Service is operational", fiber.Map{
			"service": "order-service",
			"status":  "running",
			"health":  "/api/v1/orders/health",
			"version": "1.0.0",
		})
	})

	api := app.Group("/api/v1")

	healthHandler := func(c fiber.Ctx) error {
		dbStatus := "healthy"
		if orderDB == nil || orderDB.Ping() != nil {
			dbStatus = "unreachable"
		}
		return response.Success(c, fiber.StatusOK, "Order Service is operational", fiber.Map{
			"service":   "order-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
			"version":   "1.0.0",
		})
	}
	api.Get("/health", healthHandler)
	api.Get("/orders/health", healthHandler)

	// Register Routes
	order.RegisterRoutes(api, orderHandler, cfg)
	payment.RegisterRoutes(api, paymentHandler, cfg)
	logistics.RegisterRoutes(api, logisticsHandler, cfg)

	// Graceful shutdown
	serverShutdown := make(chan os.Signal, 1)
	signal.Notify(serverShutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-serverShutdown
		log.Println("Shutting down Order Service gracefully...")
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Order Service listening on http://0.0.0.0%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("Server closed: %v", err)
	}
}
