package scheduler

import (
	"context"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type mockMetricsService struct {
	mock.Mock
}

func (m *mockMetricsService) UpdateGauge(ctx context.Context, mType string, mName string, mValue *float64) error {
	return m.Called(ctx, mType, mName, mValue).Error(0)
}

func (m *mockMetricsService) UpdateCounter(ctx context.Context, mType string, mName string, mValue *int64) error {
	return m.Called(ctx, mType, mName, mValue).Error(0)
}

func (m *mockMetricsService) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	args := m.Called(ctx, mType, mName)
	return args.String(0), args.Error(1)
}

func (m *mockMetricsService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Metrics), args.Error(1)
}

func (m *mockMetricsService) WriteMetricsFile(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockMetricsService) LoadMeticsFromFile(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestNewMetricsScheduler(t *testing.T) {
	storeInterval := int64(300)
	cfg := &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
		},
	}
	mockSvc := new(mockMetricsService)

	sched := NewMetricsScheduler(cfg, mockSvc)

	require.NotNil(t, sched)
}

func TestMetricsScheduler_RunWriteMetricsFile_StoreIntervalZero_ExitsWithoutCallingWrite(t *testing.T) {
	storeInterval := int64(0)
	cfg := &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
		},
	}
	mockSvc := new(mockMetricsService)

	mockSvc.On("WriteMetricsFile").Return(nil).Maybe()

	sched := NewMetricsScheduler(cfg, mockSvc)
	sched.RunWriteMetricsFile(t.Context())

	mockSvc.AssertNotCalled(t, "WriteMetricsFile")
}

func TestMetricsScheduler_RunWriteMetricsFile_StoreIntervalNonZero_CallsWriteMetricsFile(t *testing.T) {
	storeInterval := int64(1)
	cfg := &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
		},
	}
	mockSvc := new(mockMetricsService)
	mockSvc.On("WriteMetricsFile", mock.Anything).Return(nil).Maybe()

	sched := NewMetricsScheduler(cfg, mockSvc)

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sched.RunWriteMetricsFile(ctx)
		close(done)
	}()

	time.Sleep(1100 * time.Millisecond)

	mockSvc.AssertNumberOfCalls(t, "WriteMetricsFile", 1)
}
