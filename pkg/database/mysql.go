package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nusantara-supermart/pkg/config"
)

func ConnectMySQL(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	var db *sqlx.DB
	var err error
	for attempts := 1; attempts <= 10; attempts++ {
		db, err = sqlx.Connect("mysql", dsn)
		if err == nil {
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(10)
			db.SetConnMaxLifetime(5 * time.Minute)
			return db, nil
		}
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to MySQL after retries: %w", err)
}
