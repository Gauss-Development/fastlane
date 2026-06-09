// Package postgres provides search-service's read-only access to the catalog
// database (owned by post-service). Per the GAU-245 decision, search-service
// reads the products/suppliers tables and their embeddings directly rather than
// over gRPC, because the Product contract intentionally omits the embedding
// column. This connection must be pointed at the catalog DB and is read-only by
// convention — search-service never writes here.
package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // minutes
}

func NewConnection(cfg Config) (*sql.DB, error) {
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
