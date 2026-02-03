package service

import (
	"errors"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) UpdateGauge(metric models.Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockStorage) UpdateCounter(metric models.Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
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
				m.On("UpdateGauge", mock.MatchedBy(func(metric models.Metrics) bool {
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
				m.On("UpdateGauge", mock.AnythingOfType("models.Metrics")).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
		{
			name:   "zero value",
			mType:  models.Gauge,
			mName:  "testGauge",
			mValue: floatPtr(0.0),
			setupMock: func(m *MockStorage) {
				m.On("UpdateGauge", mock.MatchedBy(func(metric models.Metrics) bool {
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
				m.On("UpdateGauge", mock.AnythingOfType("models.Metrics")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockStorage)
			tt.setupMock(mockRepo)

			service := NewMetricsService(mockRepo)
			err := service.UpdateGauge(tt.mType, tt.mName, tt.mValue)

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
				m.On("UpdateCounter", mock.MatchedBy(func(metric models.Metrics) bool {
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
				m.On("UpdateCounter", mock.AnythingOfType("models.Metrics")).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
		{
			name:   "zero value",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(0),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.MatchedBy(func(metric models.Metrics) bool {
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
				m.On("UpdateCounter", mock.AnythingOfType("models.Metrics")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "large value",
			mType:  models.Counter,
			mName:  "testCounter",
			mValue: int64Ptr(1000000),
			setupMock: func(m *MockStorage) {
				m.On("UpdateCounter", mock.AnythingOfType("models.Metrics")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockStorage)
			tt.setupMock(mockRepo)

			service := NewMetricsService(mockRepo)
			err := service.UpdateCounter(tt.mType, tt.mName, tt.mValue)

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

	mockRepo.On("UpdateGauge", mock.MatchedBy(func(metric models.Metrics) bool {
		return metric.ID == "testGauge" &&
			metric.MType == models.Gauge &&
			metric.Value != nil &&
			*metric.Value == 123.45 &&
			metric.Hash == ""
	})).Return(nil)

	service := NewMetricsService(mockRepo)
	err := service.UpdateGauge(models.Gauge, "testGauge", floatPtr(123.45))

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMetricsService_UpdateCounter_MetricStructure(t *testing.T) {
	mockRepo := new(MockStorage)

	mockRepo.On("UpdateCounter", mock.MatchedBy(func(metric models.Metrics) bool {
		return metric.ID == "testCounter" &&
			metric.MType == models.Counter &&
			metric.Delta != nil &&
			*metric.Delta == 5 &&
			metric.Hash == ""
	})).Return(nil)

	service := NewMetricsService(mockRepo)
	err := service.UpdateCounter(models.Counter, "testCounter", int64Ptr(5))

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
