package scheduler

import (
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

func (m *mockMetricsService) UpdateGauge(string, string, *float64) error {
	return nil
}

func (m *mockMetricsService) UpdateCounter(string, string, *int64) error {
	return nil
}

func (m *mockMetricsService) GetMetric(string, string) (string, error) {
	return "", nil
}

func (m *mockMetricsService) GetAllMetrics() ([]models.Metrics, error) {
	return nil, nil
}

func (m *mockMetricsService) WriteMetricsFile() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockMetricsService) LoadMeticsFromFile() error {
	args := m.Called()
	return args.Error(0)
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
	sched.RunWriteMetricsFile()

	mockSvc.AssertNotCalled(t, "WriteMetricsFile")
}

func TestMetricsScheduler_RunWriteMetricsFile_StoreIntervalNonZero_CallsWriteMetricsFile(t *testing.T) {
	storeInterval := int64(1) // 1 секунда
	cfg := &config.Config{
		Server: config.ServerConfig{
			StoreInterval: &storeInterval,
		},
	}
	mockSvc := new(mockMetricsService)
	mockSvc.On("WriteMetricsFile").Return(nil).Maybe()

	sched := NewMetricsScheduler(cfg, mockSvc)

	done := make(chan struct{})
	go func() {
		sched.RunWriteMetricsFile()
		close(done)
	}()

	time.Sleep(1100 * time.Millisecond)

	mockSvc.AssertNumberOfCalls(t, "WriteMetricsFile", 1)
}
