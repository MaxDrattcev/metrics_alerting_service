package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
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
	_, err := p.pool.Exec(ctx, query, metric.ID, metric.MType, *metric.Value)
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

	_, err := p.pool.Exec(ctx, query, metric.ID, metric.MType, *metric.Delta)
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

	row := p.pool.QueryRow(ctx, query, mType, mName)

	var metric models.Metrics

	if err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
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
	var metrics []models.Metrics

	query := "SELECT id, type, delta, value FROM metrics"
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return []models.Metrics{}, fmt.Errorf("get metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			return []models.Metrics{}, fmt.Errorf("get metrics: %w", err)
		}
		metrics = append(metrics, metric)
	}
	err = rows.Err()
	if err != nil {
		return []models.Metrics{}, fmt.Errorf("get metrics: %w", err)
	}
	return metrics, nil
}

func (p *PostgresStorage) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("update metrics: %w", err)
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO metrics (id, type, delta, value) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET " +
		"type = EXCLUDED.type, " +
		"value = EXCLUDED.value, " +
		"delta = CASE WHEN EXCLUDED.type = 'counter' " +
		"THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0) ELSE EXCLUDED.delta END"

	for _, metric := range metrics {
		_, err := tx.Exec(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			return fmt.Errorf("update metrics: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("update metrics: %w", err)
	}
	return nil
}
