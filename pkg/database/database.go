package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config, logger *slog.Logger) (*sql.DB, error) {

	connection := fmt.Sprintf(
		" host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connection)
	if err != nil {
		logger.Error("Failed to establish connection", slog.Any("error", err))
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Error("Failed to ping database", slog.Any("error", err))
		return nil, err
	}

	return db, nil
}
