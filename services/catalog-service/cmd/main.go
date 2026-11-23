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
	"github.com/nusantara-supermart/pkg/middleware"
	"github.com/nusantara-supermart/pkg/response"
	catalogpb "github.com/nusantara-supermart/proto/catalog/v1"
	"github.com/nusantara-supermart/services/catalog-service/internal/catalog"
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
		httpPort = os.Getenv("CATALOG_SERVICE_PORT")
	}
	if httpPort == "" {
		httpPort = "3002"
	}

	grpcPort := os.Getenv("CATALOG_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	log.Printf("🚀 Starting Catalog Service (HTTP: %s, gRPC: %s) in [%s] mode...", httpPort, grpcPort, cfg.AppEnv)

	catalogDB := connectDatabase("catalog", cfg.CatalogDB)
	defer closeDatabase(catalogDB)

	catalogRepo := catalog.NewRepository(catalogDB)
	catalogSvc := catalog.NewService(catalogRepo)

	// Jalankan Catalog gRPC Server di background goroutine
	grpcServer := grpc.NewServer()
	grpcHandler := catalog.NewGRPCHandler(catalogSvc)
	catalogpb.RegisterCatalogServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for Catalog gRPC on port %s: %v", grpcPort, err)
	}

	go func() {
		log.Printf("🚀 Catalog gRPC Server listening on 0.0.0.0:%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Catalog gRPC Server stopped: %v", err)
		}
	}()

	// Inisialisasi Fiber HTTP Server
	app := fiber.New(fiber.Config{
		AppName:      "PT Nusantara SuperMart - Catalog Service v1.0",
		ServerHeader: "NusantaraSuperMart-CatalogService",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.SetupCORS(cfg))

	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "PT Nusantara SuperMart Catalog Service is operational", fiber.Map{
			"service":   "catalog-service",
			"status":    "running",
			"http_port": httpPort,
			"grpc_port": grpcPort,
			"health":    "/api/v1/catalog/health",
			"version":   "1.0.0",
		})
	})

	api := app.Group("/api/v1")

	healthHandler := func(c fiber.Ctx) error {
		dbStatus := "healthy"
		if catalogDB == nil || catalogDB.Ping() != nil {
			dbStatus = "unreachable"
		}
		return response.Success(c, fiber.StatusOK, "Catalog Service is operational", fiber.Map{
			"service":   "catalog-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
			"grpc_port": grpcPort,
			"version":   "1.0.0",
		})
	}
	api.Get("/health", healthHandler)
	api.Get("/catalog/health", healthHandler)

	catalogHandlerInstance := catalog.NewHandler(catalogSvc)
	catalog.RegisterRoutes(api, catalogHandlerInstance, cfg)

	// Graceful shutdown
	serverShutdown := make(chan os.Signal, 1)
	signal.Notify(serverShutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-serverShutdown
		log.Println("Shutting down Catalog Service gracefully...")
		grpcServer.GracefulStop()
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", httpPort)
	log.Printf("Catalog Service HTTP listening on http://0.0.0.0%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("HTTP Server closed: %v", err)
	}
}
