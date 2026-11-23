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
	"github.com/nusantara-supermart/pkg/middleware"
	"github.com/nusantara-supermart/pkg/response"
	"github.com/nusantara-supermart/services/identity-service/internal/auth"
	"github.com/nusantara-supermart/services/identity-service/internal/promotion"
	"github.com/nusantara-supermart/services/identity-service/internal/support"
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
		port = os.Getenv("IDENTITY_SERVICE_PORT")
	}
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Starting Identity Service on port %s in [%s] mode...", port, cfg.AppEnv)

	identityDB := connectDatabase("identity", cfg.IdentityDB)
	defer closeDatabase(identityDB)

	app := fiber.New(fiber.Config{
		AppName:      "PT Nusantara SuperMart - Identity Service v1.0",
		ServerHeader: "NusantaraSuperMart-IdentityService",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.SetupCORS(cfg))

	// Healthcheck & root
	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "PT Nusantara SuperMart Identity Service is operational", fiber.Map{
			"service": "identity-service",
			"status":  "running",
			"health":  "/api/v1/auth/health",
			"version": "1.0.0",
		})
	})

	api := app.Group("/api/v1")

	// Liveness / Healthcheck
	healthHandler := func(c fiber.Ctx) error {
		dbStatus := "healthy"
		if identityDB == nil || identityDB.Ping() != nil {
			dbStatus = "unreachable"
		}
		return response.Success(c, fiber.StatusOK, "Identity Service is operational", fiber.Map{
			"service":   "identity-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
			"version":   "1.0.0",
		})
	}
	api.Get("/health", healthHandler)
	api.Get("/auth/health", healthHandler)

	// Auth Domain
	authRepo := auth.NewRepository(identityDB)
	authSvc := auth.NewService(authRepo, cfg)
	authHandlerInstance := auth.NewHandler(authSvc)
	auth.RegisterRoutes(api, authHandlerInstance, cfg)

	// Promotion Domain
	promotionRepo := promotion.NewRepository(identityDB)
	promotionSvc := promotion.NewService(promotionRepo)
	promotionHandlerInstance := promotion.NewHandler(promotionSvc)
	promotion.RegisterRoutes(api, promotionHandlerInstance, cfg)

	// Support Domain
	supportRepo := support.NewRepository(identityDB)
	supportSvc := support.NewService(supportRepo)
	supportHandlerInstance := support.NewHandler(supportSvc)
	support.RegisterRoutes(api, supportHandlerInstance, cfg)

	// Graceful shutdown
	serverShutdown := make(chan os.Signal, 1)
	signal.Notify(serverShutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-serverShutdown
		log.Println("Shutting down Identity Service gracefully...")
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Identity Service listening on http://0.0.0.0%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("Server closed: %v", err)
	}
}
