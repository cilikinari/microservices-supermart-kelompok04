package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"google.golang.org/grpc"

	"github.com/nusantara-supermart/backend/internal/auth"
	"github.com/nusantara-supermart/backend/internal/catalog"
	"github.com/nusantara-supermart/backend/internal/inventory"
	inventorypb "github.com/nusantara-supermart/backend/internal/inventory/pb/inventory/v1"
	"github.com/nusantara-supermart/backend/internal/logistics"
	"github.com/nusantara-supermart/backend/internal/order"
	"github.com/nusantara-supermart/backend/internal/payment"
	"github.com/nusantara-supermart/backend/internal/procurement"
	"github.com/nusantara-supermart/backend/internal/promotion"
	"github.com/nusantara-supermart/backend/internal/support"
	"github.com/nusantara-supermart/backend/pkg/config"
	"github.com/nusantara-supermart/backend/pkg/database"
	"github.com/nusantara-supermart/backend/pkg/middleware"
	"github.com/nusantara-supermart/backend/pkg/response"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting PT Nusantara SuperMart Monolith Backend in [%s] mode...", cfg.AppEnv)

	// Database Connection
	db, err := database.ConnectMySQL(cfg)
	if err != nil {
		log.Printf("WARNING: MySQL connection failed: %v", err)
		log.Println("Backend will retry connecting when serving requests or initializing...")
	} else {
		log.Println("Connected to MySQL database successfully!")
		defer db.Close()
	}

	app := fiber.New(fiber.Config{
		AppName:      "PT Nusantara SuperMart Indonesia API v1.0",
		ServerHeader: "NusantaraSuperMart-Monolith",
	})

	// Global Middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.SetupCORS(cfg))

	// Base API Group
	api := app.Group("/api/v1")

	// Healthcheck
	api.Get("/health", func(c fiber.Ctx) error {
		dbStatus := "healthy"
		if db == nil || db.Ping() != nil {
			dbStatus = "unreachable"
		}
		return response.Success(c, fiber.StatusOK, "Nusantara SuperMart Monolith API is operational", fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
			"version":   "1.0.0",
		})
	})

	// Domain Repositories
	authRepo := auth.NewRepository(db)
	catalogRepo := catalog.NewRepository(db)
	inventoryRepo := inventory.NewRepository(db)
	orderRepo := order.NewRepository(db)
	paymentRepo := payment.NewRepository(db)
	promotionRepo := promotion.NewRepository(db)
	logisticsRepo := logistics.NewRepository(db)
	procurementRepo := procurement.NewRepository(db)
	supportRepo := support.NewRepository(db)

	// Inisialisasi gRPC Client untuk Inventory (Digunakan oleh Order)
	grpcInventoryClient := order.NewInventoryClient("localhost:50051")

	// Domain Services
	authSvc := auth.NewService(authRepo, cfg)
	catalogSvc := catalog.NewService(catalogRepo)
	inventorySvc := inventory.NewService(inventoryRepo)
	// Passing grpcInventoryClient ke Order Service
	orderSvc := order.NewService(orderRepo, catalogSvc, inventorySvc, grpcInventoryClient)
	paymentSvc := payment.NewService(paymentRepo, orderSvc)
	promotionSvc := promotion.NewService(promotionRepo)
	logisticsSvc := logistics.NewService(logisticsRepo, orderSvc)
	procurementSvc := procurement.NewService(procurementRepo)
	supportSvc := support.NewService(supportRepo)

	// Domain Handlers (REST / Fiber)
	authHandler := auth.NewHandler(authSvc)
	catalogHandler := catalog.NewHandler(catalogSvc)
	inventoryHandler := inventory.NewHandler(inventorySvc)
	orderHandler := order.NewHandler(orderSvc)
	paymentHandler := payment.NewHandler(paymentSvc)
	promotionHandler := promotion.NewHandler(promotionSvc)
	logisticsHandler := logistics.NewHandler(logisticsSvc)
	procurementHandler := procurement.NewHandler(procurementSvc)
	supportHandler := support.NewHandler(supportSvc)

	// Register Domain Routers (REST / Fiber)
	auth.RegisterRoutes(api, authHandler, cfg)
	catalog.RegisterRoutes(api, catalogHandler, cfg)
	inventory.RegisterRoutes(api, inventoryHandler, cfg)
	order.RegisterRoutes(api, orderHandler, cfg)
	payment.RegisterRoutes(api, paymentHandler, cfg)
	promotion.RegisterRoutes(api, promotionHandler, cfg)
	logistics.RegisterRoutes(api, logisticsHandler, cfg)
	procurement.RegisterRoutes(api, procurementHandler, cfg)
	support.RegisterRoutes(api, supportHandler, cfg)

	// Jalankan gRPC Server (Inventory) di Port 50051
	go func() {
		grpcPort := ":50051"
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC on %s: %v", grpcPort, err)
		}

		grpcServer := grpc.NewServer()

		// Register Inventory gRPC Handler
		grpcInventoryHandler := inventory.NewGRPCHandler(inventorySvc)
		inventorypb.RegisterInventoryServiceServer(grpcServer, grpcInventoryHandler)

		log.Printf("🚀 gRPC Server listening on 0.0.0.0%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Graceful Shutdown Setup
	serverShutdown := make(chan os.Signal, 1)
	signal.Notify(serverShutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-serverShutdown
		log.Println("Shutting down Nusantara SuperMart server gracefully...")
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Nusantara SuperMart server listening on http://0.0.0.0%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("Server closed: %v", err)
	}
}