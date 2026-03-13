package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) UpdateGauge(ctx context.Context, metric models.Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockStorage) UpdateCounter(ctx context.Context, metric models.Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockStorage) GetMetric(ctx context.Context, mType string, mName string) (models.Metrics, error) {
	args := m.Called(ctx, mType, mName)
	if args.Get(0) == nil {
		return models.Metrics{}, args.Error(1)
	}
	return args.Get(0).(models.Metrics), args.Error(1)
}

func (m *MockStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Metrics), args.Error(1)
}

type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) WriteMetrics(metrics []models.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *MockFileStorage) ReadMetrics() ([]models.Metrics, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Metrics), args.Error(1)
}

func floatPtr(v float64) *float64 { return &v }
func int64Ptr(v int64) *int64     { return &v }

func testServerConfig(storeInterval int64, restore bool) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
			Restore:       &restore,
		},
	}
}

func TestMetricsService_UpdateGauge(t *testing.T) {
	tests := []struct {
		name      string
		mType     string
		mName     string
		mValue    *float64
		setupMock func(storage *MockStorage)
		wantErr   bool
	}{
		{
			name:   "successful update",
			mType:  models.Gauge,
			mName:  "testGauge",
			mValue: floatPtr(123.45),
			setupMock: func(m *MockStorage) {
				m.On("UpdateGauge", mock.Anything, mock.MatchedBy(func(metric models.Metrics) bool {
					return metric.ID == "testGauge" &&
						metric.MType == models.Gauge &&
						metric.Value != nil &&
						*metric.Value == 123.45
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			mType:  models.Gauge,
			mName:  "testGauge",
			mValue: floatPtr(123.45),
			setupMock: func(m *MockStorage) {
				m.On("UpdateGauge", mock.Anything, mock.AnythingOfType("models.Metrics")).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
		{
			name:   "zero value",
			mType:  models.Gauge,
			mName:  "testGauge",
			mValue: floatPtr(0.0),
			setupMock: func(m *MockStorage) {
				m.On("UpdateGauge", mock.Anything, mock.MatchedBy(func(metric models.Metrics) bool {
					return metric.Value != nil && *metric.Value == 0.0
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "negative value",
			mType:  models.Gauge,
			mName:  "testGauge",
			mValue: floatPtr(-100.5),
			setupMock: func(m *MockStorage) {
				m.On("UpdateGauge", mock.Anything, mock.AnythingOfType("models.Metrics")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockStorage)
			mockFile := new(MockFileStorage)
			cfg := testServerConfig(300, false)
			tt.setupMock(mockRepo)

			svc := NewMetricsService(mockRepo, mockFile, cfg)
			err := svc.UpdateGauge(context.Background(), tt.mType, tt.mName, tt.mValue)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricsService_UpdateCounter(t *testing.T) {
	tests := []struct {
		name      string
		mType     string
		mName     string
		mValue    *int64
		setupMock func(*MockStorage)
		wantErr   bool
	}{
		{
			name:   "successful update",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(5),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.Anything, mock.MatchedBy(func(metric models.Metrics) bool {
					return metric.ID == "testCounter" &&
						metric.MType == models.Counter &&
						metric.Delta != nil &&
						*metric.Delta == 5
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(5),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.Anything, mock.AnythingOfType("models.Metrics")).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
		{
			name:   "zero value",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(0),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.Anything, mock.MatchedBy(func(metric models.Metrics) bool {
					return metric.Delta != nil && *metric.Delta == 0
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "negative value",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(-10),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.Anything, mock.AnythingOfType("models.Metrics")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "large value",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(1000000),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.Anything, mock.AnythingOfType("models.Metrics")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockStorage)
			mockFile := new(MockFileStorage)
			cfg := testServerConfig(300, false)
			tt.setupMock(mockRepo)

			svc := NewMetricsService(mockRepo, mockFile, cfg)
			err := svc.UpdateCounter(context.Background(), tt.mType, tt.mName, tt.mValue)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricsService_UpdateGauge_MetricStructure(t *testing.T) {
	mockRepo := new(MockStorage)
	mockFile := new(MockFileStorage)
	cfg := testServerConfig(300, false)

	mockRepo.On("UpdateGauge", mock.Anything, mock.MatchedBy(func(metric models.Metrics) bool {
		return metric.ID == "testGauge" &&
			metric.MType == models.Gauge &&
			metric.Value != nil &&
			*metric.Value == 123.45 &&
			metric.Hash == ""
	})).Return(nil)

	svc := NewMetricsService(mockRepo, mockFile, cfg)
	err := svc.UpdateGauge(context.Background(), models.Gauge, "testGauge", floatPtr(123.45))

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMetricsService_UpdateCounter_MetricStructure(t *testing.T) {
	mockRepo := new(MockStorage)
	mockFile := new(MockFileStorage)
	cfg := testServerConfig(300, false)

	mockRepo.On("UpdateCounter", mock.Anything, mock.MatchedBy(func(metric models.Metrics) bool {
		return metric.ID == "testCounter" &&
			metric.MType == models.Counter &&
			metric.Delta != nil &&
			*metric.Delta == 5 &&
			metric.Hash == ""
	})).Return(nil)

	svc := NewMetricsService(mockRepo, mockFile, cfg)
	err := svc.UpdateCounter(context.Background(), models.Counter, "testCounter", int64Ptr(5))

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMetricsService_GetMetric(t *testing.T) {
	tests := []struct {
		name      string
		mType     string
		mName     string
		setupMock func(*MockStorage)
		wantValue string
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "successful get gauge metric",
			mType: models.Gauge,
			mName: "testGauge",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Gauge, "testGauge").Return(
					models.Metrics{
						ID:    "testGauge",
						MType: models.Gauge,
						Value: floatPtr(123.45),
					},
					nil,
				)
			},
			wantValue: "123.45",
			wantErr:   false,
		},
		{
			name:  "successful get counter metric",
			mType: models.Counter,
			mName: "testCounter",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Counter, "testCounter").Return(
					models.Metrics{
						ID:    "testCounter",
						MType: models.Counter,
						Delta: int64Ptr(5),
					},
					nil,
				)
			},
			wantValue: "5",
			wantErr:   false,
		},
		{
			name:  "gauge with zero value",
			mType: models.Gauge,
			mName: "zeroGauge",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Gauge, "zeroGauge").Return(
					models.Metrics{
						ID:    "zeroGauge",
						MType: models.Gauge,
						Value: floatPtr(0.0),
					},
					nil,
				)
			},
			wantValue: "0",
			wantErr:   false,
		},
		{
			name:  "counter with zero value",
			mType: models.Counter,
			mName: "zeroCounter",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Counter, "zeroCounter").Return(
					models.Metrics{
						ID:    "zeroCounter",
						MType: models.Counter,
						Delta: int64Ptr(0),
					},
					nil,
				)
			},
			wantValue: "0",
			wantErr:   false,
		},
		{
			name:  "gauge with negative value",
			mType: models.Gauge,
			mName: "negativeGauge",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Gauge, "negativeGauge").Return(
					models.Metrics{
						ID:    "negativeGauge",
						MType: models.Gauge,
						Value: floatPtr(-123.45),
					},
					nil,
				)
			},
			wantValue: "-123.45",
			wantErr:   false,
		},
		{
			name:  "counter with large value",
			mType: models.Counter,
			mName: "largeCounter",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Counter, "largeCounter").Return(
					models.Metrics{
						ID:    "largeCounter",
						MType: models.Counter,
						Delta: int64Ptr(1000000),
					},
					nil,
				)
			},
			wantValue: "1000000",
			wantErr:   false,
		},
		{
			name:  "metric not found",
			mType: models.Gauge,
			mName: "nonExistent",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Gauge, "nonExistent").Return(
					models.Metrics{},
					fmt.Errorf("metric not found"),
				)
			},
			wantValue: "",
			wantErr:   true,
			errMsg:    "metric not found",
		},
		{
			name:  "gauge with nil value",
			mType: models.Gauge,
			mName: "nilGauge",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Gauge, "nilGauge").Return(
					models.Metrics{
						ID:    "nilGauge",
						MType: models.Gauge,
						Value: nil,
					},
					nil,
				)
			},
			wantValue: "",
			wantErr:   true,
			errMsg:    "gauge metric value is nil",
		},
		{
			name:  "counter with nil delta",
			mType: models.Counter,
			mName: "nilCounter",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Counter, "nilCounter").Return(
					models.Metrics{
						ID:    "nilCounter",
						MType: models.Counter,
						Delta: nil,
					},
					nil,
				)
			},
			wantValue: "",
			wantErr:   true,
			errMsg:    "counter metric delta is nil",
		},
		{
			name:  "gauge with decimal precision",
			mType: models.Gauge,
			mName: "decimalGauge",
			setupMock: func(m *MockStorage) {
				m.On("GetMetric", mock.Anything, models.Gauge, "decimalGauge").Return(
					models.Metrics{
						ID:    "decimalGauge",
						MType: models.Gauge,
						Value: floatPtr(123.456789),
					},
					nil,
				)
			},
			wantValue: "123.456789",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockStorage)
			mockFile := new(MockFileStorage)
			cfg := testServerConfig(300, false)
			tt.setupMock(mockRepo)

			svc := NewMetricsService(mockRepo, mockFile, cfg)
			value, err := svc.GetMetric(context.Background(), tt.mType, tt.mName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantValue, value)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricsService_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockStorage)
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful get all metrics",
			setupMock: func(m *MockStorage) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{
						{ID: "gauge1", MType: models.Gauge, Value: floatPtr(123.45)},
						{ID: "counter1", MType: models.Counter, Delta: int64Ptr(5)},
					},
					nil,
				)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "empty metrics list",
			setupMock: func(m *MockStorage) {
				m.On("GetAllMetrics", mock.Anything).Return([]models.Metrics{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single gauge metric",
			setupMock: func(m *MockStorage) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{
						{ID: "gauge1", MType: models.Gauge, Value: floatPtr(123.45)},
					},
					nil,
				)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "single counter metric",
			setupMock: func(m *MockStorage) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{
						{ID: "counter1", MType: models.Counter, Delta: int64Ptr(5)},
					},
					nil,
				)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple metrics",
			setupMock: func(m *MockStorage) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{
						{ID: "gauge1", MType: models.Gauge, Value: floatPtr(123.45)},
						{ID: "gauge2", MType: models.Gauge, Value: floatPtr(67.89)},
						{ID: "counter1", MType: models.Counter, Delta: int64Ptr(5)},
						{ID: "counter2", MType: models.Counter, Delta: int64Ptr(10)},
					},
					nil,
				)
			},
			wantCount: 4,
			wantErr:   false,
		},
		{
			name: "repository error",
			setupMock: func(m *MockStorage) {
				m.On("GetAllMetrics", mock.Anything).Return(nil, fmt.Errorf("repository error"))
			},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "repository error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockStorage)
			mockFile := new(MockFileStorage)
			cfg := testServerConfig(300, false)
			tt.setupMock(mockRepo)

			svc := NewMetricsService(mockRepo, mockFile, cfg)
			metrics, err := svc.GetAllMetrics(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				require.Nil(t, metrics)
			} else {
				require.NoError(t, err)
				require.NotNil(t, metrics)
				require.Equal(t, tt.wantCount, len(metrics))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricsService_WriteMetricsFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, false)

		metrics := []models.Metrics{
			{ID: "g1", MType: models.Gauge, Value: floatPtr(1.0)},
		}
		mockRepo.On("GetAllMetrics", mock.Anything).Return(metrics, nil)
		mockFile.On("WriteMetrics", metrics).Return(nil)

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.WriteMetricsFile(context.Background())

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockFile.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, false)

		mockRepo.On("GetAllMetrics", mock.Anything).Return(nil, errors.New("repo error"))

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.WriteMetricsFile(context.Background())

		require.Error(t, err)
		require.Contains(t, err.Error(), "repo error")
		mockRepo.AssertExpectations(t)
		mockFile.AssertNotCalled(t, "WriteMetrics", mock.Anything)
	})

	t.Run("file write error", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, false)

		mockRepo.On("GetAllMetrics", mock.Anything).Return([]models.Metrics{}, nil)
		mockFile.On("WriteMetrics", mock.Anything).Return(errors.New("write error"))

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.WriteMetricsFile(context.Background())

		require.Error(t, err)
		require.Contains(t, err.Error(), "write error")
		mockRepo.AssertExpectations(t)
		mockFile.AssertExpectations(t)
	})
}

func TestMetricsService_LoadMeticsFromFile(t *testing.T) {
	t.Run("restore false returns nil without reading file", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, false)

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.LoadMeticsFromFile(context.Background())

		require.NoError(t, err)
		mockFile.AssertNotCalled(t, "ReadMetrics")
		mockRepo.AssertExpectations(t)
	})

	t.Run("restore true success", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, true)

		metrics := []models.Metrics{
			{ID: "g1", MType: models.Gauge, Value: floatPtr(1.0)},
			{ID: "c1", MType: models.Counter, Delta: int64Ptr(10)},
		}
		mockFile.On("ReadMetrics").Return(metrics, nil)
		mockRepo.On("UpdateGauge", mock.Anything, mock.MatchedBy(func(m models.Metrics) bool { return m.ID == "g1" })).Return(nil)
		mockRepo.On("UpdateCounter", mock.Anything, mock.MatchedBy(func(m models.Metrics) bool { return m.ID == "c1" })).Return(nil)

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.LoadMeticsFromFile(context.Background())

		require.NoError(t, err)
		mockFile.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("restore true file read error", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, true)

		mockFile.On("ReadMetrics").Return(nil, errors.New("read error"))

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.LoadMeticsFromFile(context.Background())

		require.Error(t, err)
		require.Contains(t, err.Error(), "read error")
		mockFile.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "UpdateGauge", mock.Anything, mock.Anything)
		mockRepo.AssertNotCalled(t, "UpdateCounter", mock.Anything, mock.Anything)
	})

	t.Run("restore true empty file", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, true)

		mockFile.On("ReadMetrics").Return([]models.Metrics{}, nil)

		svc := NewMetricsService(mockRepo, mockFile, cfg)
		err := svc.LoadMeticsFromFile(context.Background())

		require.NoError(t, err)
		mockFile.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "UpdateGauge", mock.Anything, mock.Anything)
		mockRepo.AssertNotCalled(t, "UpdateCounter", mock.Anything, mock.Anything)
	})
}
