package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"net/http"
)

const (
	contentType = "Content-Type"
	textType    = "text/plain"
	jsonType    = "application/json"
)

type MetricsSender struct {
	cfg    *config.Config
	client *RetryableClient
}

func NewMetricsSender(cfg *config.Config) *MetricsSender {
	return &MetricsSender{
		client: NewRetryableClient(),
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

func (s *MetricsSender) SendGaugeJSON(ctx context.Context, name string, value float64) error {
	metric := models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}
	return s.sendMetricJSONGzip(ctx, metric)
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

func (s *MetricsSender) SendCounterJSON(ctx context.Context, name string, value int64) error {
	metric := models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
	}
	return s.sendMetricJSONGzip(ctx, metric)
}

func (s *MetricsSender) sendMetricJSONGzip(ctx context.Context, metric models.Metrics) error {
	payload, err := json.Marshal(&metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}
	buf, err := s.compressGzip(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/update", s.cfg.Client.Address)
	headers := map[string]string{
		contentType:        jsonType,
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}

	if s.cfg.Client.Key != "" {
		hash, err := s.computeHashSHA256(payload)
		if err != nil {
			return fmt.Errorf("failed compute hash: %w", err)
		}
		headers["HashSHA256"] = hash
	}

	resp, err := s.client.PostWithRetry(ctx, url, headers, buf.Bytes())
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (s *MetricsSender) SendAllMetricsBuffer(ctx context.Context, metrics []models.Metrics) error {
	payload, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	buf, err := s.compressGzip(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/updates", s.cfg.Client.Address)
	headers := map[string]string{
		contentType:        jsonType,
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}
	if s.cfg.Client.Key != "" {
		hash, err := s.computeHashSHA256(payload)
		if err != nil {
			return fmt.Errorf("failed compute hash: %w", err)
		}
		headers["HashSHA256"] = hash
	}
	resp, err := s.client.PostWithRetry(ctx, url, headers, buf.Bytes())
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (s *MetricsSender) compressGzip(payload []byte) (bytes.Buffer, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(payload); err != nil {
		gz.Close()
		return bytes.Buffer{}, fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return bytes.Buffer{}, fmt.Errorf("gzip close: %w", err)
	}
	return buf, nil
}

func (s *MetricsSender) computeHashSHA256(bodyBytes []byte) (string, error) {
	mac := hmac.New(sha256.New, []byte(s.cfg.Client.Key))
	mac.Write(bodyBytes)
	return hex.EncodeToString(mac.Sum(nil)), nil
}
