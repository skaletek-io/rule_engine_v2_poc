package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pgxpool from DATABASE_URL env var.
// Expected format: postgres://user:pass@host:port/dbname?sslmode=disable
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	return pgxpool.New(ctx, dsn)
}
