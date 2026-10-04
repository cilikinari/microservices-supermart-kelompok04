package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppPort     string
	IdentityDB  DatabaseConfig
	CatalogDB   DatabaseConfig
	InventoryDB DatabaseConfig
	JWTSecret   string
	JWTExpiry   string
	CORSOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	appEnv := getEnv("APP_ENV", "development")
	appPort := getEnv("APP_PORT", "3000")
	identityDB := loadDatabaseConfig("IDENTITY_DB", "127.0.0.1", "3306", "identity_service_db", "identity_admin", "identity_secret_pass")
	catalogDB := loadDatabaseConfig("CATALOG_DB", "127.0.0.1", "3307", "catalog_service_db", "catalog_admin", "catalog_secret_pass")
	inventoryDB := loadDatabaseConfig("INVENTORY_DB", "127.0.0.1", "3308", "inventory_service_db", "inventory_admin", "inventory_secret_pass")
	jwtSecret := getEnv("JWT_SECRET", "supermart-secret-key-for-nusantara-course-2026")
	jwtExpiry := getEnv("JWT_EXPIRY", "24h")
	corsRaw := getEnv("CORS_ORIGINS", "http://localhost:4200,http://127.0.0.1:4200,http://localhost:3000")

	corsOrigins := strings.Split(corsRaw, ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	return &Config{
		AppEnv:      appEnv,
		AppPort:     appPort,
		IdentityDB:  identityDB,
		CatalogDB:   catalogDB,
		InventoryDB: inventoryDB,
		JWTSecret:   jwtSecret,
		JWTExpiry:   jwtExpiry,
		CORSOrigins: corsOrigins,
	}
}

func loadDatabaseConfig(prefix, host, port, database, user, password string) DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv(prefix+"_HOST", host),
		Port:     getEnv(prefix+"_PORT", port),
		Database: getEnv(prefix+"_NAME", database),
		User:     getEnv(prefix+"_USER", user),
		Password: getEnv(prefix+"_PASSWORD", password),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
