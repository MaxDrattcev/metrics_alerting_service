package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// exampleServer поднимает тестовый HTTP-сервер с маршрутами практического трека.
func exampleServer() *httptest.Server {
	dir, err := os.MkdirTemp("", "metrics-example-*")
	if err != nil {
		return nil
	}

	storeInterval := int64(300)
	restore := false
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address:         "localhost:8080",
			FileStoragePath: filepath.Join(dir, "metrics.json"),
			StoreInterval:   &storeInterval,
			Restore:         &restore,
		},
	}

	repo := repository.NewMemStorage()
	file := repository.NewFileStorage(cfg.Server.FileStoragePath)
	svc := service.NewMetricsService(repo, file, cfg, nil)

	legacy := NewMetricsHandler(svc)
	jsonH := NewMetricsJSONHandler(svc, cfg)

	r := gin.New()
	r.POST("/update/:type/:name/:value", legacy.Update)
	r.GET("/value/:type/:name", legacy.GetMetric)
	r.POST("/update", jsonH.Update)
	r.POST("/update/", jsonH.Update)
	r.POST("/value", jsonH.GetMetric)
	r.POST("/value/", jsonH.GetMetric)
	r.GET("/metrics", jsonH.GetAllMetrics)
	r.POST("/updates", jsonH.UpdateMetrics)
	r.POST("/updates/", jsonH.UpdateMetrics)

	return httptest.NewServer(r)
}

// ExampleNewMetricsHandler_updateGauge — POST /update/{type}/{name}/{value} (текстовый API).
func ExampleNewMetricsHandler_updateGauge() {
	ts := exampleServer()
	if ts == nil {
		return
	}
	defer ts.Close()

	resp, err := http.Post(
		ts.URL+"/update/gauge/Alloc/27",
		"text/plain",
		http.NoBody,
	)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
	// Output:
	// 200
	//
}

// ExampleNewMetricsHandler_getGauge — GET /value/{type}/{name} (текстовый API).
func ExampleNewMetricsHandler_getGauge() {
	ts := exampleServer()
	if ts == nil {
		return
	}
	defer ts.Close()

	setupResp, err := http.Post(ts.URL+"/update/gauge/Alloc/27", "text/plain", http.NoBody)
	if err != nil {
		return
	}
	setupResp.Body.Close()

	resp, err := http.Get(ts.URL + "/value/gauge/Alloc")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
	// Output:
	// 200
	// 27
}

// ExampleNewMetricsJSONHandler_update — POST /update (JSON, одна метрика).
func ExampleNewMetricsJSONHandler_update() {
	ts := exampleServer()
	if ts == nil {
		return
	}
	defer ts.Close()

	value := 42.5
	payload, err := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &value,
	})
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
	// Output:
	// 200
	//
}

// ExampleNewMetricsJSONHandler_getValue — POST /value (JSON: запрос и ответ с актуальным значением).
func ExampleNewMetricsJSONHandler_getValue() {
	ts := exampleServer()
	if ts == nil {
		return
	}
	defer ts.Close()

	value := 42.5
	updateBody, err := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &value,
	})
	if err != nil {
		return
	}

	reqUpdate, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", bytes.NewReader(updateBody))
	if err != nil {
		return
	}
	reqUpdate.Header.Set("Content-Type", "application/json")

	updateResp, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		return
	}
	updateResp.Body.Close()

	getBody, err := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
	})
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/value/", bytes.NewReader(getBody))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":42.5}
}

// ExampleNewMetricsJSONHandler_updates — POST /updates (пакет метрик).
func ExampleNewMetricsJSONHandler_updates() {
	ts := exampleServer()
	if ts == nil {
		return
	}
	defer ts.Close()

	v1, v2 := 1.0, 2.0
	metrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &v1},
		{ID: "HeapAlloc", MType: models.Gauge, Value: &v2},
	}
	payload, err := json.Marshal(metrics)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/updates/", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
	// Output:
	// 200
	//
}

// ExampleNewMetricsJSONHandler_getAllMetrics — GET /metrics (все метрики в JSON).
func ExampleNewMetricsJSONHandler_getAllMetrics() {
	ts := exampleServer()
	if ts == nil {
		return
	}
	defer ts.Close()

	value := 10.0
	payload, err := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &value,
	})
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	updateResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	updateResp.Body.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output:
	// 200
	// [{"id":"Alloc","type":"gauge","value":10}]
}
