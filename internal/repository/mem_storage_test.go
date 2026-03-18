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
			storage := NewMemStorage()
			err := storage.UpdateGauge(t.Context(), tt.metric)

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
			storage := NewMemStorage()
			err := storage.UpdateCounter(t.Context(), test.metric)
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
			storage := NewMemStorage()

			for _, metric := range test.updates {
				err := storage.UpdateCounter(t.Context(), metric)
				if test.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*MemStorage) // Функция для подготовки данных
		mType     string
		mName     string
		wantErr   bool
		errMsg    string
		wantValue *float64 // Для gauge
		wantDelta *int64   // Для counter
	}{
		{
			name: "get existing gauge metric",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "testGauge",
					MType: models.Gauge,
					Value: floatPtr(123.45),
				})
			},
			mType:     models.Gauge,
			mName:     "testGauge",
			wantErr:   false,
			wantValue: floatPtr(123.45),
		},
		{
			name: "get existing counter metric",
			setup: func(s *MemStorage) {
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "testCounter",
					MType: models.Counter,
					Delta: int64Ptr(5),
				})
			},
			mType:     models.Counter,
			mName:     "testCounter",
			wantErr:   false,
			wantDelta: int64Ptr(5),
		},
		{
			name:    "get non-existent metric",
			setup:   func(s *MemStorage) {},
			mType:   models.Gauge,
			mName:   "nonExistent",
			wantErr: true,
			errMsg:  "metric not found",
		},
		{
			name: "get gauge with zero value",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "zeroGauge",
					MType: models.Gauge,
					Value: floatPtr(0.0),
				})
			},
			mType:     models.Gauge,
			mName:     "zeroGauge",
			wantErr:   false,
			wantValue: floatPtr(0.0),
		},
		{
			name: "get counter with zero value",
			setup: func(s *MemStorage) {
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "zeroCounter",
					MType: models.Counter,
					Delta: int64Ptr(0),
				})
			},
			mType:     models.Counter,
			mName:     "zeroCounter",
			wantErr:   false,
			wantDelta: int64Ptr(0),
		},
		{
			name: "get counter with accumulated value",
			setup: func(s *MemStorage) {
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "accumCounter",
					MType: models.Counter,
					Delta: int64Ptr(5),
				})
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "accumCounter",
					MType: models.Counter,
					Delta: int64Ptr(3),
				})
			},
			mType:     models.Counter,
			mName:     "accumCounter",
			wantErr:   false,
			wantDelta: int64Ptr(8), // 5 + 3
		},
		{
			name: "wrong type for existing name",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "testMetric",
					MType: models.Gauge,
					Value: floatPtr(123.45),
				})
			},
			mType:   models.Counter, // Ищем counter, но есть gauge
			mName:   "testMetric",
			wantErr: true,
			errMsg:  "metric not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage().(*MemStorage)
			tt.setup(storage)

			metric, err := storage.GetMetric(t.Context(), tt.mType, tt.mName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mName, metric.ID)
				assert.Equal(t, tt.mType, metric.MType)

				if tt.mType == models.Gauge {
					require.NotNil(t, metric.Value)
					assert.Equal(t, *tt.wantValue, *metric.Value)
				} else if tt.mType == models.Counter {
					require.NotNil(t, metric.Delta)
					assert.Equal(t, *tt.wantDelta, *metric.Delta)
				}
			}
		})
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*MemStorage)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty storage",
			setup:     func(s *MemStorage) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single gauge metric",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "gauge1",
					MType: models.Gauge,
					Value: floatPtr(123.45),
				})
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "single counter metric",
			setup: func(s *MemStorage) {
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(5),
				})
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple metrics",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "gauge1",
					MType: models.Gauge,
					Value: floatPtr(123.45),
				})
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "gauge2",
					MType: models.Gauge,
					Value: floatPtr(67.89),
				})
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "counter1",
					MType: models.Counter,
					Delta: int64Ptr(5),
				})
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "multiple metrics with same name different types",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "testMetric",
					MType: models.Gauge,
					Value: floatPtr(123.45),
				})
				s.UpdateCounter(t.Context(), models.Metrics{
					ID:    "testMetric",
					MType: models.Counter,
					Delta: int64Ptr(5),
				})
			},
			wantCount: 2, // Разные типы = разные ключи
			wantErr:   false,
		},
		{
			name: "overwrite gauge metric",
			setup: func(s *MemStorage) {
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "gauge1",
					MType: models.Gauge,
					Value: floatPtr(100.0),
				})
				s.UpdateGauge(t.Context(), models.Metrics{
					ID:    "gauge1",
					MType: models.Gauge,
					Value: floatPtr(200.0),
				})
			},
			wantCount: 1, // Перезаписано, остается одна метрика
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage().(*MemStorage)
			tt.setup(storage)

			metrics, err := storage.GetAllMetrics(t.Context())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(metrics))

				for _, metric := range metrics {
					assert.NotEmpty(t, metric.ID)
					assert.True(t, metric.MType == models.Gauge || metric.MType == models.Counter)
					if metric.MType == models.Gauge {
						assert.NotNil(t, metric.Value)
					} else if metric.MType == models.Counter {
						assert.NotNil(t, metric.Delta)
					}
				}
			}
		})
	}
}
