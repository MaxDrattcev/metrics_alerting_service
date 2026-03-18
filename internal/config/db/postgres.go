package db

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func NewConDB(ctx context.Context, cfg config.Config, pathMigration string) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	if cfg.Server.DatabaseDSN != "" {
		config, err := pgxpool.ParseConfig(cfg.Server.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("parse dsn: %w", err)
		}

		config.MaxConns = 10
		config.MinConns = 2
		config.MaxConnLifetime = time.Hour
		config.MaxConnIdleTime = 30 * time.Minute

		var err1 error
		pool, err1 = pgxpool.NewWithConfig(ctx, config)
		if err1 != nil {
			return nil, fmt.Errorf("create pool: %w", err1)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, fmt.Errorf("ping db: %w", err)
		}

		m, err := migrate.New(pathMigration, cfg.Server.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("migrate init: %w", err)
		}
		defer m.Close()

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, fmt.Errorf("migrate up: %w", err)
		}
	}
	return pool, nil
}
