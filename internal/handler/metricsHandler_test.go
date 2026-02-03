package handler

import (
	"errors"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) UpdateGauge(mType string, mName string, mValue *float64) error {
	args := m.Called(mType, mName, mValue)
	return args.Error(0)
}

func (m *MockService) UpdateCounter(mType string, mName string, mValue *int64) error {
	args := m.Called(mType, mName, mValue)
	return args.Error(0)
}

func TestMetricsHandler_Update_Gauge(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		contentType string
		setupMock   func(*MockService)
		wantStatus  int
		wantErr     bool
	}{
		{
			name:        "successful gauge update",
			method:      http.MethodPost,
			path:        "/update/gauge/testGauge/123.45",
			contentType: "text/plain",
			setupMock: func(m *MockService) {
				m.On("UpdateGauge", models.Gauge, "testGauge", mock.AnythingOfType("*float64")).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:        "invalid method GET",
			method:      http.MethodGet,
			path:        "/update/gauge/testGauge/123.45",
			contentType: "text/plain",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusMethodNotAllowed,
			wantErr:     false,
		},
		{
			name:        "invalid Content-Type",
			method:      http.MethodPost,
			path:        "/update/gauge/testGauge/123.45",
			contentType: "application/json",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusBadRequest,
			wantErr:     false,
		},
		{
			name:        "invalid path format",
			method:      http.MethodPost,
			path:        "/update/gauge/testGauge",
			contentType: "text/plain",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusNotFound,
			wantErr:     false,
		},
		{
			name:        "empty metric name",
			method:      http.MethodPost,
			path:        "/update/gauge//123.45",
			contentType: "text/plain",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusNotFound,
			wantErr:     false,
		},
		{
			name:        "invalid gauge value",
			method:      http.MethodPost,
			path:        "/update/gauge/testGauge/invalid",
			contentType: "text/plain",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusBadRequest,
			wantErr:     false,
		},
		{
			name:        "service error",
			method:      http.MethodPost,
			path:        "/update/gauge/testGauge/123.45",
			contentType: "text/plain",
			setupMock: func(m *MockService) {
				m.On("UpdateGauge", models.Gauge, "testGauge", mock.AnythingOfType("*float64")).Return(errors.New("service error"))
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.setupMock(mockService)

			handler := NewMetricsHandler(mockService)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			handler.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricsHandler_Update_Counter(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		contentType string
		setupMock   func(*MockService)
		wantStatus  int
	}{
		{
			name:        "successful counter update",
			method:      http.MethodPost,
			path:        "/update/counter/testCounter/5",
			contentType: "text/plain",
			setupMock: func(m *MockService) {
				m.On("UpdateCounter", models.Counter, "testCounter", mock.AnythingOfType("*int64")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "invalid counter value",
			method:      http.MethodPost,
			path:        "/update/counter/testCounter/invalid",
			contentType: "text/plain",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty metric name",
			method:      http.MethodPost,
			path:        "/update/counter//5",
			contentType: "text/plain",
			setupMock:   func(m *MockService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "service error",
			method:      http.MethodPost,
			path:        "/update/counter/testCounter/5",
			contentType: "text/plain",
			setupMock: func(m *MockService) {
				m.On("UpdateCounter", models.Counter, "testCounter", mock.AnythingOfType("*int64")).Return(errors.New("service error"))
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.setupMock(mockService)

			handler := NewMetricsHandler(mockService)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			handler.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricsHandler_Update_InvalidMetricType(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		contentType string
		wantStatus  int
	}{
		{
			name:        "invalid metric type",
			path:        "/update/invalid/testMetric/123",
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty metric type",
			path:        "/update//testMetric/123",
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewMetricsHandler(mockService)

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			handler.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
