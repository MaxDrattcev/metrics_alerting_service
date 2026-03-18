package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) MetricsStorage {
	return &PostgresStorage{
		pool: pool,
	}
}

func (p *PostgresStorage) UpdateGauge(ctx context.Context, metric models.Metrics) error {
	if metric.Value == nil {
		return fmt.Errorf("gauge metric value is nil")
	}
	query := "INSERT INTO metrics (id, type, delta, value) VALUES ($1, $2, NULL, $3) " +
		"ON CONFLICT (id) DO UPDATE SET type = EXCLUDED.type, value = EXCLUDED.value, delta = NULL"

	err := WithTxRetry(ctx, p.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, metric.ID, metric.MType, *metric.Value)
		return err
	})
	if err != nil {
		return fmt.Errorf("update gauge: %w", err)
	}
	return nil
}

func (p *PostgresStorage) UpdateCounter(ctx context.Context, metric models.Metrics) error {
	if metric.Delta == nil {
		return fmt.Errorf("counter metric delta is nil")
	}
	query := "INSERT INTO metrics (id, type, delta, value) VALUES ($1, $2, $3, NULL)" +
		"ON CONFLICT (id) DO UPDATE SET type = EXCLUDED.type, value = NULL, delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta"

	err := WithTxRetry(ctx, p.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, metric.ID, metric.MType, *metric.Delta)
		return err
	})
	if err != nil {
		return fmt.Errorf("update counter: %w", err)
	}
	return nil
}

func (p *PostgresStorage) GetMetric(ctx context.Context, mType string, mName string) (models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return models.Metrics{}, err
	}
	query := "SELECT * FROM metrics WHERE type = $1 AND id = $2"

	var metric models.Metrics
	err := WithTxRetry(ctx, p.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, mType, mName)
		return row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Metrics{}, fmt.Errorf("metric not found")
		}
		return models.Metrics{}, fmt.Errorf("get metric: %w", err)
	}

	return metric, nil
}

func (p *PostgresStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return []models.Metrics{}, err
	}

	query := "SELECT id, type, delta, value FROM metrics"

	var out []models.Metrics
	err := WithTxRetry(ctx, p.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		out = out[:0]

		for rows.Next() {
			var m models.Metrics
			if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
				return err
			}
			out = append(out, m)
		}
		return rows.Err()
	})
	if err != nil {
		return []models.Metrics{}, fmt.Errorf("get metrics: %w", err)
	}

	return out, nil
}

func (p *PostgresStorage) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	query := "INSERT INTO metrics (id, type, delta, value) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET " +
		"type = EXCLUDED.type, " +
		"value = EXCLUDED.value, " +
		"delta = CASE WHEN EXCLUDED.type = 'counter' " +
		"THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0) ELSE EXCLUDED.delta END"

	err := WithTxRetry(ctx, p.pool, func(tx pgx.Tx) error {
		for _, metric := range metrics {
			if _, err := tx.Exec(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update metrics: %w", err)
	}
	return nil
}
