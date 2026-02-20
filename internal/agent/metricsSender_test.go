package agent

import (
	"encoding/json"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMetricsSender_SendGauge(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue float64
		serverFunc  func(http.ResponseWriter, *http.Request)
		wantErr     bool
	}{
		{
			name:        "successful send",
			metricName:  "testGauge",
			metricValue: 123.45,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
				assert.Equal(t, "/update/gauge/testGauge/123.45", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "server error 500",
			metricName:  "testGauge",
			metricValue: 123.45,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name:        "server error 400",
			metricName:  "testGauge",
			metricValue: 123.45,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
		},
		{
			name:        "zero value",
			metricName:  "testGauge",
			metricValue: 0.0,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/gauge/testGauge/0", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "negative value",
			metricName:  "testGauge",
			metricValue: -100.5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/gauge/testGauge/-100.5", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый HTTP сервер
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			cfg := &config.Config{
				Client: config.ClientConfig{
					Address: serverURL.Host,
				},
			}

			sender := NewMetricsSender(cfg)
			err = sender.SendGauge(tt.metricName, tt.metricValue)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricsSender_SendCounter(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue int64
		serverFunc  func(http.ResponseWriter, *http.Request)
		wantErr     bool
	}{
		{
			name:        "successful send",
			metricName:  "testCounter",
			metricValue: 5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
				assert.Equal(t, "/update/counter/testCounter/5", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "server error 500",
			metricName:  "testCounter",
			metricValue: 5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name:        "zero value",
			metricName:  "testCounter",
			metricValue: 0,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/counter/testCounter/0", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "negative value",
			metricName:  "testCounter",
			metricValue: -10,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/counter/testCounter/-10", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "large value",
			metricName:  "testCounter",
			metricValue: 1000000,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/counter/testCounter/1000000", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			cfg := &config.Config{
				Client: config.ClientConfig{
					Address: serverURL.Host,
				},
			}

			sender := NewMetricsSender(cfg)
			err = sender.SendCounter(tt.metricName, tt.metricValue)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricsSender_SendGauge_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		Client: config.ClientConfig{
			Address: "invalid-host:9999",
		},
	}

	sender := NewMetricsSender(cfg)
	err := sender.SendGauge("testGauge", 123.45)

	require.Error(t, err, "Should return error for invalid URL")
}

func TestMetricsSender_SendCounter_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		Client: config.ClientConfig{
			Address: "invalid-host:9999",
		},
	}

	sender := NewMetricsSender(cfg)
	err := sender.SendCounter("testCounter", 5)

	require.Error(t, err, "Should return error for invalid URL")
}

func TestMetricsSender_sendGaugeJson(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue float64
		serverFunc  func(http.ResponseWriter, *http.Request)
		wantErr     bool
	}{
		{
			name:        "successful send",
			metricName:  "testGauge",
			metricValue: 123.45,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "/update", r.URL.Path)
				var m models.Metrics
				err := json.NewDecoder(r.Body).Decode(&m)
				require.NoError(t, err)
				assert.Equal(t, "testGauge", m.ID)
				assert.Equal(t, models.Gauge, m.MType)
				require.NotNil(t, m.Value)
				assert.Equal(t, 123.45, *m.Value)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "server error 500",
			metricName:  "testGauge",
			metricValue: 123.45,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name:        "server error 400",
			metricName:  "testGauge",
			metricValue: 123.45,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
		},
		{
			name:        "zero value",
			metricName:  "testGauge",
			metricValue: 0.0,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				var m models.Metrics
				require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
				require.NotNil(t, m.Value)
				assert.Equal(t, 0.0, *m.Value)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "negative value",
			metricName:  "testGauge",
			metricValue: -100.5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				var m models.Metrics
				require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
				require.NotNil(t, m.Value)
				assert.Equal(t, -100.5, *m.Value)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			cfg := &config.Config{
				Client: config.ClientConfig{
					Address: serverURL.Host,
				},
			}

			sender := NewMetricsSender(cfg)
			err = sender.sendGaugeJson(tt.metricName, tt.metricValue)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricsSender_sendCounterJson(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue int64
		serverFunc  func(http.ResponseWriter, *http.Request)
		wantErr     bool
	}{
		{
			name:        "successful send",
			metricName:  "testCounter",
			metricValue: 5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "/update", r.URL.Path)
				var m models.Metrics
				err := json.NewDecoder(r.Body).Decode(&m)
				require.NoError(t, err)
				assert.Equal(t, "testCounter", m.ID)
				assert.Equal(t, models.Counter, m.MType)
				require.NotNil(t, m.Delta)
				assert.Equal(t, int64(5), *m.Delta)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "server error 500",
			metricName:  "testCounter",
			metricValue: 5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name:        "server error 400",
			metricName:  "testCounter",
			metricValue: 5,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
		},
		{
			name:        "zero value",
			metricName:  "testCounter",
			metricValue: 0,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				var m models.Metrics
				require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
				require.NotNil(t, m.Delta)
				assert.Equal(t, int64(0), *m.Delta)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "negative value",
			metricName:  "testCounter",
			metricValue: -10,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				var m models.Metrics
				require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
				require.NotNil(t, m.Delta)
				assert.Equal(t, int64(-10), *m.Delta)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:        "large value",
			metricName:  "testCounter",
			metricValue: 1000000,
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				var m models.Metrics
				require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
				require.NotNil(t, m.Delta)
				assert.Equal(t, int64(1000000), *m.Delta)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			cfg := &config.Config{
				Client: config.ClientConfig{
					Address: serverURL.Host,
				},
			}

			sender := NewMetricsSender(cfg)
			err = sender.sendCounterJson(tt.metricName, tt.metricValue)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricsSender_sendGaugeJson_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		Client: config.ClientConfig{
			Address: "invalid-host:9999",
		},
	}
	sender := NewMetricsSender(cfg)
	err := sender.sendGaugeJson("testGauge", 123.45)
	require.Error(t, err)
}

func TestMetricsSender_sendCounterJson_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		Client: config.ClientConfig{
			Address: "invalid-host:9999",
		},
	}
	sender := NewMetricsSender(cfg)
	err := sender.sendCounterJson("testCounter", 5)
	require.Error(t, err)
}
