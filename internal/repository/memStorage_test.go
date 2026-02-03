package repository

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	tests := []struct {
		name    string
		metric  models.Metrics
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid gauge metric",
			metric: models.Metrics{
				ID:    "testGauge",
				MType: models.Gauge,
				Value: floatPtr(123.45),
			},
			wantErr: false,
		}, {
			name: "nil gauge metric",
			metric: models.Metrics{
				ID:    "testGauge",
				MType: models.Gauge,
				Value: nil,
			},
			wantErr: true,
			errMsg:  "gauge metric requires value",
		}, {
			name: "zero value is valid",
			metric: models.Metrics{
				ID:    "testGauge",
				MType: models.Gauge,
				Value: floatPtr(0.0),
			},
			wantErr: false,
		}, {
			name: "negative value is valid",
			metric: models.Metrics{
				ID:    "testGauge",
				MType: models.Gauge,
				Value: floatPtr(-123.45),
			},
			wantErr: false,
		}, {
			name: "small value is valid",
			metric: models.Metrics{
				ID:    "testGauge",
				MType: models.Gauge,
				Value: floatPtr(0.0001),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMetricsStorage()
			err := storage.UpdateGauge(tt.metric)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	tests := []struct {
		name    string
		metric  models.Metrics
		wantErr bool
	}{
		{
			name: "valid counter metric",
			metric: models.Metrics{
				ID:    "testCounter",
				MType: models.Counter,
				Delta: int64Ptr(5),
			},
			wantErr: false,
		}, {
			name: "zero value is valid",
			metric: models.Metrics{
				ID:    "testCounter",
				MType: models.Counter,
				Delta: int64Ptr(0),
			},
			wantErr: false,
		}, {
			name: "negative value is valid",
			metric: models.Metrics{
				ID:    "testCounter",
				MType: models.Counter,
				Delta: int64Ptr(-10),
			},
			wantErr: false,
		}, {
			name: "large value is valid",
			metric: models.Metrics{
				ID:    "testCounter",
				MType: models.Counter,
				Delta: int64Ptr(100000000),
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMetricsStorage()
			err := storage.UpdateCounter(test.metric)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMemStorage_UpdateCounter_Summation(t *testing.T) {
	tests := []struct {
		name    string
		updates []models.Metrics
		wantErr bool
	}{
		{
			name: "single update",
			updates: []models.Metrics{
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(5),
				},
			},
			wantErr: false,
		},
		{
			name: "two updates - should sum",
			updates: []models.Metrics{
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(5),
				},
				{
					ID:    "counter2",
					MType: models.Counter,
					Delta: int64Ptr(3),
				},
			},
			wantErr: false,
		},
		{
			name: "three updates - should sum",
			updates: []models.Metrics{
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(5),
				},
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(3),
				},
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(2),
				},
			},
			wantErr: false,
		}, {
			name: "sum with zero",
			updates: []models.Metrics{
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(0),
				},
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(3),
				},
			},
			wantErr: false,
		}, {
			name: "sum with negative",
			updates: []models.Metrics{
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(-10),
				},
				{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(1000),
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMetricsStorage()

			for _, metric := range test.updates {
				err := storage.UpdateCounter(metric)
				if test.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}
