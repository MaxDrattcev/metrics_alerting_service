package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/mocks"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsJSONHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		method      string
		contentType string
		body        interface{}
		setupMock   func(*mocks.MockMetricsService)
		wantStatus  int
	}{
		{
			name:        "successful gauge update",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "g1", "type": "gauge", "value": 123.45},
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateGauge", mock.Anything, models.Gauge, "g1", mock.AnythingOfType("*float64")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "successful counter update",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "c1", "type": "counter", "delta": int64(10)},
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateCounter", mock.Anything, models.Counter, "c1", mock.AnythingOfType("*int64")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "wrong Content-Type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        map[string]interface{}{"id": "g1", "type": "gauge", "value": 1.0},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "invalid JSON body",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{invalid`,
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty id",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "", "type": "gauge", "value": 1.0},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "invalid type",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "x", "type": "unknown", "value": 1.0},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "gauge without value",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "g1", "type": "gauge"},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "counter without delta",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "c1", "type": "counter"},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "service error on UpdateGauge",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "g1", "type": "gauge", "value": 1.0},
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateGauge", mock.Anything, models.Gauge, "g1", mock.AnythingOfType("*float64")).Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewMockMetricsService(t)
			tt.setupMock(mockSvc)
			h := NewMetricsJSONHandler(mockSvc)

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			case map[string]interface{}:
				bodyBytes, _ = json.Marshal(b)
			default:
				bodyBytes, _ = json.Marshal(tt.body)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, "/update", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", tt.contentType)

			h.Update(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestMetricsJSONHandler_GetMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		method      string
		contentType string
		body        interface{}
		setupMock   func(*mocks.MockMetricsService)
		wantStatus  int
		checkBody   func(t *testing.T, body []byte)
	}{
		{
			name:        "successful get gauge",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "g1", "type": "gauge"},
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Gauge, "g1").Return("123.45", nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var m models.Metrics
				require.NoError(t, json.Unmarshal(body, &m))
				require.NotNil(t, m.Value)
				assert.Equal(t, 123.45, *m.Value)
				assert.Equal(t, "g1", m.ID)
				assert.Equal(t, models.Gauge, m.MType)
			},
		},
		{
			name:        "successful get counter",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "c1", "type": "counter"},
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Counter, "c1").Return("42", nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var m models.Metrics
				require.NoError(t, json.Unmarshal(body, &m))
				require.NotNil(t, m.Delta)
				assert.Equal(t, int64(42), *m.Delta)
			},
		},
		{
			name:        "wrong Content-Type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        map[string]interface{}{"id": "g1", "type": "gauge"},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "invalid JSON",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{`,
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty id",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "", "type": "gauge"},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "invalid type",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "x", "type": "unknown"},
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "service error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        map[string]interface{}{"id": "g1", "type": "gauge"},
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Gauge, "g1").Return("", errors.New("not found"))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewMockMetricsService(t)
			tt.setupMock(mockSvc)
			h := NewMetricsJSONHandler(mockSvc)

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			case map[string]interface{}:
				bodyBytes, _ = json.Marshal(b)
			default:
				bodyBytes, _ = json.Marshal(tt.body)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, "/value", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", tt.contentType)

			h.GetMetric(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.checkBody != nil && w.Code == http.StatusOK {
				tt.checkBody(t, w.Body.Bytes())
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestMetricsJSONHandler_GetAllMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		setupMock  func(*mocks.MockMetricsService)
		wantStatus int
	}{
		{
			name:   "successful get all",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return([]models.Metrics{
					{ID: "a", MType: models.Gauge, Value: floatPtr(1.0)},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "service error",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := mocks.NewMockMetricsService(t)
			tt.setupMock(mockSvc)
			h := NewMetricsJSONHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, "/", nil)

			h.GetAllMetrics(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
