package handler

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"net/http"
	"strconv"
	"strings"
)

const incorrectValue = "Incorrect metric value"

type metricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(service service.MetricsService) MetricsHandler {
	return &metricsHandler{
		service: service,
	}
}
func (m *metricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !m.validateRequest(w, r) {
		return
	}

	mType, mName, mValue, ok := m.parsePath(w, r.URL.Path)
	if !ok {
		return
	}

	if !m.validateParam(w, mType, mName, mValue) {
		return
	}

	if mType == models.Gauge {
		floatVal, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			http.Error(w, incorrectValue, http.StatusBadRequest)
			return
		}

		if err := m.service.UpdateGauge(mType, mName, &floatVal); err != nil {
			http.Error(w, incorrectValue, http.StatusBadRequest)
			return
		}
	}

	if mType == models.Counter {
		intVal, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			http.Error(w, incorrectValue, http.StatusBadRequest)
			return
		}
		if err := m.service.UpdateCounter(mType, mName, &intVal); err != nil {
			http.Error(w, incorrectValue, http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (m *metricsHandler) validateRequest(w http.ResponseWriter, r *http.Request) bool {
	method := r.Method
	if method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	contentHeader := r.Header.Get("Content-Type")

	if contentHeader != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return false
	}
	return true
}

func (m *metricsHandler) parsePath(w http.ResponseWriter, path string) (mType, mName, mValue string, ok bool) {
	path = strings.TrimPrefix(path, "/update/")
	parts := strings.Split(path, "/")

	mType = ""
	mName = ""
	mValue = ""

	if len(parts) > 0 {
		mType = parts[0]
	}
	if len(parts) > 1 {
		mName = parts[1]
	}
	if len(parts) > 2 {
		mValue = parts[2]
	}

	if mType == "" {
		http.Error(w, "Incorrect metric type", http.StatusBadRequest)
		return "", "", "", false
	}

	return mType, mName, mValue, true
}

func (m *metricsHandler) validateParam(w http.ResponseWriter, mType, mName, mValue string) bool {

	if mName == "" {
		http.Error(w, "Name cannot be empty", http.StatusNotFound)
		return false
	}

	if mType == "" || mType != models.Gauge && mType != models.Counter {
		http.Error(w, "Incorrect metric type", http.StatusBadRequest)
		return false
	}

	if mValue == "" {
		http.Error(w, "Value cannot be empty", http.StatusNotFound)
		return false
	}
	return true
}
