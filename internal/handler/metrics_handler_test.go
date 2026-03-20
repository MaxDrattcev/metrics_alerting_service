package handler

import (
	"errors"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/mocks"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestMetricsHandler_Update_Gauge(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		typeParam   string
		nameParam   string
		valueParam  string
		contentType string
		setupMock   func(*mocks.MockMetricsService)
		wantStatus  int
		wantErr     bool
	}{
		{
			name:        "successful gauge update",
			method:      http.MethodPost,
			typeParam:   "gauge",
			nameParam:   "testGauge",
			valueParam:  "123.45",
			contentType: "text/plain",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateGauge", mock.Anything, models.Gauge, "testGauge", mock.AnythingOfType("*float64")).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:        "empty metric name",
			method:      http.MethodPost,
			typeParam:   "gauge",
			nameParam:   "",
			valueParam:  "123.45",
			contentType: "text/plain",
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusNotFound,
			wantErr:     false,
		},
		{
			name:        "invalid gauge value",
			method:      http.MethodPost,
			typeParam:   "gauge",
			nameParam:   "testGauge",
			valueParam:  "invalid",
			contentType: "text/plain",
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusBadRequest,
			wantErr:     false,
		},
		{
			name:        "service error",
			method:      http.MethodPost,
			typeParam:   "gauge",
			nameParam:   "testGauge",
			valueParam:  "123.45",
			contentType: "text/plain",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateGauge", mock.Anything, models.Gauge, "testGauge", mock.AnythingOfType("*float64")).Return(errors.New("service error"))
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockMetricsService(t)
			tt.setupMock(mockService)

			handler := NewMetricsHandler(mockService)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(tt.method, "/update/"+tt.typeParam+"/"+tt.nameParam+"/"+tt.valueParam, nil)
			req.Header.Set("Content-Type", tt.contentType)
			c.Request = req

			c.Params = gin.Params{
				{Key: "type", Value: tt.typeParam},
				{Key: "name", Value: tt.nameParam},
				{Key: "value", Value: tt.valueParam},
			}

			handler.Update(c)

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
		typeParam   string
		nameParam   string
		valueParam  string
		contentType string
		setupMock   func(*mocks.MockMetricsService)
		wantStatus  int
	}{
		{
			name:        "successful counter update",
			method:      http.MethodPost,
			typeParam:   "counter",
			nameParam:   "testCounter",
			valueParam:  "5",
			contentType: "text/plain",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateCounter", mock.Anything, models.Counter, "testCounter", mock.AnythingOfType("*int64")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "invalid counter value",
			method:      http.MethodPost,
			typeParam:   "counter",
			nameParam:   "testCounter",
			valueParam:  "invalid",
			contentType: "text/plain",
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty metric name",
			method:      http.MethodPost,
			typeParam:   "counter",
			nameParam:   "",
			valueParam:  "5",
			contentType: "text/plain",
			setupMock:   func(*mocks.MockMetricsService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "service error",
			method:      http.MethodPost,
			typeParam:   "counter",
			nameParam:   "testCounter",
			valueParam:  "5",
			contentType: "text/plain",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("UpdateCounter", mock.Anything, models.Counter, "testCounter", mock.AnythingOfType("*int64")).Return(errors.New("service error"))
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockMetricsService(t)
			tt.setupMock(mockService)

			handler := NewMetricsHandler(mockService)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(tt.method, "/update/"+tt.typeParam+"/"+tt.nameParam+"/"+tt.valueParam, nil)
			req.Header.Set("Content-Type", tt.contentType)
			c.Request = req

			c.Params = gin.Params{
				{Key: "type", Value: tt.typeParam},
				{Key: "name", Value: tt.nameParam},
				{Key: "value", Value: tt.valueParam},
			}

			handler.Update(c)

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
		typeParam   string
		nameParam   string
		valueParam  string
		contentType string
		wantStatus  int
	}{
		{
			name:        "invalid metric type",
			typeParam:   "invalid",
			nameParam:   "testMetric",
			valueParam:  "123",
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty metric type",
			typeParam:   "",
			nameParam:   "testMetric",
			valueParam:  "123",
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockMetricsService(t)
			handler := NewMetricsHandler(mockService)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodPost, "/update/"+tt.typeParam+"/"+tt.nameParam+"/"+tt.valueParam, nil)
			req.Header.Set("Content-Type", tt.contentType)
			c.Request = req

			c.Params = gin.Params{
				{Key: "type", Value: tt.typeParam},
				{Key: "name", Value: tt.nameParam},
				{Key: "value", Value: tt.valueParam},
			}

			handler.Update(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMetricsHandler_GetMetric(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		typeParam  string
		nameParam  string
		setupMock  func(*mocks.MockMetricsService)
		wantStatus int
		wantBody   string
	}{
		{
			name:      "successful get gauge metric",
			method:    http.MethodGet,
			typeParam: "gauge",
			nameParam: "testGauge",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Gauge, "testGauge").Return("123.45", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "123.45",
		},
		{
			name:      "successful get counter metric",
			method:    http.MethodGet,
			typeParam: "counter",
			nameParam: "testCounter",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Counter, "testCounter").Return("5", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "5",
		},
		{
			name:       "empty metric type",
			method:     http.MethodGet,
			typeParam:  "",
			nameParam:  "testGauge",
			setupMock:  func(*mocks.MockMetricsService) {},
			wantStatus: http.StatusNotFound,
			wantBody:   "Type cannot be empty",
		},
		{
			name:       "empty metric name",
			method:     http.MethodGet,
			typeParam:  "gauge",
			nameParam:  "",
			setupMock:  func(*mocks.MockMetricsService) {},
			wantStatus: http.StatusNotFound,
			wantBody:   "Name cannot be empty",
		},
		{
			name:       "invalid metric type",
			method:     http.MethodGet,
			typeParam:  "invalid",
			nameParam:  "testMetric",
			setupMock:  func(*mocks.MockMetricsService) {},
			wantStatus: http.StatusNotFound,
			wantBody:   incorrectType,
		},
		{
			name:      "metric not found",
			method:    http.MethodGet,
			typeParam: "gauge",
			nameParam: "nonExistent",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Gauge, "nonExistent").Return("", errors.New("metric not found"))
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "metric not found",
		},
		{
			name:      "service error",
			method:    http.MethodGet,
			typeParam: "gauge",
			nameParam: "testGauge",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Gauge, "testGauge").Return("", errors.New("service error"))
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "service error",
		},
		{
			name:      "gauge with zero value",
			method:    http.MethodGet,
			typeParam: "gauge",
			nameParam: "zeroGauge",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Gauge, "zeroGauge").Return("0", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "0",
		},
		{
			name:      "counter with zero value",
			method:    http.MethodGet,
			typeParam: "counter",
			nameParam: "zeroCounter",
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetMetric", mock.Anything, models.Counter, "zeroCounter").Return("0", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockMetricsService(t)
			tt.setupMock(mockService)

			handler := NewMetricsHandler(mockService)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(tt.method, "/value/"+tt.typeParam+"/"+tt.nameParam, nil)
			c.Request = req

			c.Params = gin.Params{
				{Key: "type", Value: tt.typeParam},
				{Key: "name", Value: tt.nameParam},
			}

			handler.GetMetric(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		setupMock  func(*mocks.MockMetricsService)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "successful get all metrics",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{
						{ID: "gauge1", MType: models.Gauge, Value: floatPtr(123.45)},
						{ID: "counter1", MType: models.Counter, Delta: int64Ptr(5)},
					},
					nil,
				)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "empty metrics list",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return([]models.Metrics{}, nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "service error",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return(nil, errors.New("service error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    false,
		},
		{
			name:   "single gauge metric",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{{ID: "gauge1", MType: models.Gauge, Value: floatPtr(123.45)}},
					nil,
				)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "single counter metric",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{{ID: "counter1", MType: models.Counter, Delta: int64Ptr(5)}},
					nil,
				)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "metrics with nil values",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
				m.On("GetAllMetrics", mock.Anything).Return(
					[]models.Metrics{
						{ID: "nilGauge", MType: models.Gauge, Value: nil},
						{ID: "nilCounter", MType: models.Counter, Delta: nil},
					},
					nil,
				)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "multiple metrics",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockMetricsService) {
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
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockMetricsService(t)
			tt.setupMock(mockService)

			handler := NewMetricsHandler(mockService)

			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.SetHTMLTemplate(template.Must(template.New("metrics.html").Parse(`
<!DOCTYPE html>
<html>
<head><title>Metrics</title></head>
<body>
    <h1>Metrics</h1>
    <table>
        <thead>
            <tr><th>Name</th><th>Type</th><th>Value</th></tr>
        </thead>
        <tbody>
            {{range .metrics}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Type}}</td>
                <td>{{.Value}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
			`)))

			router.Any("/", handler.GetAllMetrics)

			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
				assert.NotEmpty(t, w.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
