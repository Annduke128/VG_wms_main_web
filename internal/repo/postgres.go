package repo

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	Pool *pgxpool.Pool
}

func NewPostgresRepo(ctx context.Context) (*PostgresRepo, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://wms:wms_secret@localhost:5432/wms?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	config.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &PostgresRepo{Pool: pool}, nil
}

func (r *PostgresRepo) Close() {
	r.Pool.Close()
}
