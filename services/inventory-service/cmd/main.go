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
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/nusantara-supermart/pkg/config"
	"github.com/nusantara-supermart/pkg/database"
	"github.com/nusantara-supermart/pkg/messaging"
	"github.com/nusantara-supermart/pkg/middleware"
	"github.com/nusantara-supermart/pkg/response"
	inventorypb "github.com/nusantara-supermart/proto/inventory/v1"
	"github.com/nusantara-supermart/services/inventory-service/internal/inventory"
	"github.com/nusantara-supermart/services/inventory-service/internal/procurement"
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

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = os.Getenv("INVENTORY_SERVICE_PORT")
	}
	if httpPort == "" {
		httpPort = "3003"
	}

	grpcPort := os.Getenv("INVENTORY_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	catalogGRPCAddr := os.Getenv("CATALOG_GRPC_ADDR")
	if catalogGRPCAddr == "" {
		// Jika berjalan di docker compose gunakan nama service catalog-service, fallback localhost
		if os.Getenv("APP_ENV") == "production" || os.Getenv("DOCKER_CONTAINER") != "" {
			catalogGRPCAddr = "catalog-service:50052"
		} else {
			catalogGRPCAddr = "localhost:50052"
		}
	}

	log.Printf("🚀 Starting Inventory Service (HTTP: %s, gRPC: %s) in [%s] mode...", httpPort, grpcPort, cfg.AppEnv)

	inventoryDB := connectDatabase("inventory", cfg.InventoryDB)
	defer closeDatabase(inventoryDB)

	inventoryRepo := inventory.NewRepository(inventoryDB)
	inventorySvc := inventory.NewService(inventoryRepo)

	// Inisialisasi gRPC Client ke Catalog (Catalog Upstream/Supplier -> Inventory Downstream/Customer)
	catalogGRPCClient := inventory.NewCatalogClient(catalogGRPCAddr)
	if catalogGRPCClient != nil {
		defer catalogGRPCClient.Close()
		inventorySvc.SetCatalogClient(catalogGRPCClient)
		log.Printf("Inventory connected to Catalog gRPC at %s successfully", catalogGRPCAddr)
	}

	// Inisialisasi Real Consumer untuk domain Inventory
	inventoryConsumer := inventory.NewInventoryConsumer(inventoryDB)

	// Koneksi RabbitMQ & jalankan Consumer
	rabbitMQ, err := messaging.ConnectRabbitMQ()
	if err != nil {
		log.Printf("WARNING: Gagal koneksi ke RabbitMQ: %v. Berjalan dalam mode fallback.", err)
	} else {
		defer rabbitMQ.Close()
		log.Println("RabbitMQ connected successfully in Inventory Service")
		if err := rabbitMQ.ConsumeOrderEvents(inventoryConsumer.HandleOrderPlaced, inventoryDB); err != nil {
			log.Printf("WARNING: Gagal menjalankan consumer RabbitMQ: %v", err)
		} else {
			log.Println("RabbitMQ Inventory Consumer started successfully")
		}
	}

	// Jalankan Inventory gRPC Server di background goroutine
	grpcServer := grpc.NewServer()
	grpcHandler := inventory.NewGRPCHandler(inventorySvc)
	inventorypb.RegisterInventoryServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for Inventory gRPC on port %s: %v", grpcPort, err)
	}

	go func() {
		log.Printf("🚀 Inventory gRPC Server listening on 0.0.0.0:%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Inventory gRPC Server stopped: %v", err)
		}
	}()

	// Inisialisasi Fiber HTTP Server
	app := fiber.New(fiber.Config{
		AppName:      "PT Nusantara SuperMart - Inventory Service v1.0",
		ServerHeader: "NusantaraSuperMart-InventoryService",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.SetupCORS(cfg))

	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "PT Nusantara SuperMart Inventory Service is operational", fiber.Map{
			"service":   "inventory-service",
			"status":    "running",
			"http_port": httpPort,
			"grpc_port": grpcPort,
			"health":    "/api/v1/inventory/health",
			"version":   "1.0.0",
		})
	})

	api := app.Group("/api/v1")

	healthHandler := func(c fiber.Ctx) error {
		dbStatus := "healthy"
		if inventoryDB == nil || inventoryDB.Ping() != nil {
			dbStatus = "unreachable"
		}
		return response.Success(c, fiber.StatusOK, "Inventory Service is operational", fiber.Map{
			"service":   "inventory-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
			"grpc_port": grpcPort,
			"version":   "1.0.0",
		})
	}
	api.Get("/health", healthHandler)
	api.Get("/inventory/health", healthHandler)

	// Routes
	inventoryHandlerInstance := inventory.NewHandler(inventorySvc)
	inventory.RegisterRoutes(api, inventoryHandlerInstance, cfg)

	// Procurement Domain
	procurementRepo := procurement.NewRepository(inventoryDB)
	procurementSvc := procurement.NewService(procurementRepo)
	procurementHandlerInstance := procurement.NewHandler(procurementSvc)
	procurement.RegisterRoutes(api, procurementHandlerInstance, cfg)

	// Graceful shutdown
	serverShutdown := make(chan os.Signal, 1)
	signal.Notify(serverShutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-serverShutdown
		log.Println("Shutting down Inventory Service gracefully...")
		grpcServer.GracefulStop()
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", httpPort)
	log.Printf("Inventory Service HTTP listening on http://0.0.0.0%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("HTTP Server closed: %v", err)
	}
}
