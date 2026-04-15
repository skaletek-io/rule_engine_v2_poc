package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/sql/*.sql
var migrationFiles embed.FS

// RunMigrations applies all pending up-migrations against the given pool.
// It is safe to call on every startup — already-applied migrations are skipped.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	src, err := iofs.New(migrationFiles, "migrations/sql")
	if err != nil {
		return fmt.Errorf("migrate: load source: %w", err)
	}

	// golang-migrate pgx/v5 driver expects the "pgx5://" scheme.
	dsn := pool.Config().ConnString()
	pgxDSN := pgx5DSN(dsn)

	m, err := migrate.NewWithSourceInstance("iofs", src, pgxDSN)
	if err != nil {
		return fmt.Errorf("migrate: init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: up: %w", err)
	}

	version, _, _ := m.Version()
	log.Printf("migrations: at version %d", version)
	return nil
}

// pgx5DSN rewrites postgres:// or postgresql:// to pgx5://.
func pgx5DSN(dsn string) string {
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if strings.HasPrefix(dsn, prefix) {
			return "pgx5://" + dsn[len(prefix):]
		}
	}
	return dsn
}
