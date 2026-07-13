package postgres

import (
	"database/sql"
	"errors"
	"time"

	"catalog-service/internal/config"

	_ "github.com/lib/pq"
)

// ErrNoRows is a sentinel re-exported so callers can avoid importing database/sql.
var ErrNoRows = sql.ErrNoRows

func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
