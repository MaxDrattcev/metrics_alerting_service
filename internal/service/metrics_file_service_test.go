package service

import (
	"context"
	"errors"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

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
func TestMetricsFileService_WriteMetricsFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, false)

		metrics := []models.Metrics{
			{ID: "g1", MType: models.Gauge, Value: floatPtr(1.0)},
		}
		mockRepo.On("GetAllMetrics", mock.Anything).Return(metrics, nil)
		mockFile.On("WriteMetrics", metrics).Return(nil)

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
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

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
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

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
		err := svc.WriteMetricsFile(context.Background())

		require.Error(t, err)
		require.Contains(t, err.Error(), "write error")
		mockRepo.AssertExpectations(t)
		mockFile.AssertExpectations(t)
	})
}

func TestMetricsFileService_LoadMeticsFromFile(t *testing.T) {
	t.Run("restore false returns nil without reading file", func(t *testing.T) {
		mockRepo := new(MockStorage)
		mockFile := new(MockFileStorage)
		cfg := testServerConfig(300, false)

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
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

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
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

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
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

		svc := NewMetricsFileService(mockRepo, mockFile, cfg)
		err := svc.LoadMeticsFromFile(context.Background())

		require.NoError(t, err)
		mockFile.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "UpdateGauge", mock.Anything, mock.Anything)
		mockRepo.AssertNotCalled(t, "UpdateCounter", mock.Anything, mock.Anything)
	})
}
