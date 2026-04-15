package store

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/sqlc"
)

// Store wraps sqlc Queries with the underlying connection pool.
type Store struct {
	*sqlc.Queries
	Pool *pgxpool.Pool
}

// New returns a Store backed by pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: sqlc.New(pool),
		Pool:    pool,
	}
}
