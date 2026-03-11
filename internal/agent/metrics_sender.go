package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

const (
	contentType = "Content-Type"
	textType    = "text/plain"
	jsonType    = "application/json"
)

type MetricsSender struct {
	cfg    *config.Config
	client *resty.Client
}

func NewMetricsSender(cfg *config.Config) *MetricsSender {
	client := resty.New()
	client.SetTimeout(5 * time.Second)

	return &MetricsSender{
		client: client,
		cfg:    cfg,
	}
}

func (s *MetricsSender) SendGauge(name string, value float64) error {
	url := fmt.Sprintf("http://%s/update/gauge/%s/%v",
		s.cfg.Client.Address, name, value)

	response, err := s.client.R().
		SetHeader(contentType, textType).
		Post(url)
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return nil
}

func (s *MetricsSender) SendGaugeJSON(name string, value float64) error {
	metric := models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}
	return s.sendMetricJSONGzip(metric)
}

func (s *MetricsSender) SendCounter(name string, value int64) error {
	url := fmt.Sprintf("http://%s/update/counter/%s/%d",
		s.cfg.Client.Address, name, value)

	response, err := s.client.R().
		SetHeader(contentType, textType).
		Post(url)
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return nil
}

func (s *MetricsSender) SendCounterJSON(name string, value int64) error {
	metric := models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
	}
	return s.sendMetricJSONGzip(metric)
}

func (s *MetricsSender) sendMetricJSONGzip(metric models.Metrics) error {
	payload, err := json.Marshal(&metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(payload); err != nil {
		gz.Close()
		return fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close: %w", err)
	}

	url := fmt.Sprintf("http://%s/update", s.cfg.Client.Address)
	resp, err := s.client.R().
		SetHeader(contentType, jsonType).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
