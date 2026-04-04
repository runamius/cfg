package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pool.Ping: %w", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("runMigrations: %w", err)
	}

	return pool, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationFile := "schema/000001_init.up.sql"
	data, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("reading migration file %s: %w", migrationFile, err)
	}

	if _, err := pool.Exec(ctx, string(data)); err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}

	return nil
}
