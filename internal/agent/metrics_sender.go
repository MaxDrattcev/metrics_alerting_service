package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/crypto"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/hasher"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
)

const (
	contentType = "Content-Type"
	textType    = "text/plain"
	jsonType    = "application/json"
)

// MetricsSender отправляет метрики на сервер по HTTP (gzip, RSA, batch).
type MetricsSender struct {
	cfg       *config.Config
	client    *RetryableClient
	publicKey *rsa.PublicKey
}

// NewMetricsSender создаёт отправитель метрик.
func NewMetricsSender(cfg *config.Config) (*MetricsSender, error) {
	s := &MetricsSender{
		client: NewRetryableClient(),
		cfg:    cfg,
	}

	if cfg.Client.CryptoKey == "" {
		return s, nil
	}

	pub, err := crypto.LoadPublicKey(cfg.Client.CryptoKey)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}
	s.publicKey = pub
	return s, nil
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

	body, err := s.prepareRequestBody(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/update", s.cfg.Client.Address)
	headers := map[string]string{
		contentType:        jsonType,
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}
	if ip := s.hostIP(); ip != "" {
		headers["X-Real-IP"] = ip
	}

	if s.cfg.Client.Key != "" {
		hash, err := hasher.ComputeHashSHA256(payload, s.cfg.Client.Key)
		if err != nil {
			return fmt.Errorf("failed compute hash: %w", err)
		}
		headers["HashSHA256"] = hash
	}

	resp, err := s.client.PostWithRetry(ctx, url, headers, body)
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

	body, err := s.prepareRequestBody(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/updates", s.cfg.Client.Address)
	headers := map[string]string{
		contentType:        jsonType,
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}
	if ip := s.hostIP(); ip != "" {
		headers["X-Real-IP"] = ip
	}
	if s.cfg.Client.Key != "" {
		hash, err := hasher.ComputeHashSHA256(payload, s.cfg.Client.Key)
		if err != nil {
			return fmt.Errorf("failed compute hash: %w", err)
		}
		headers["HashSHA256"] = hash
	}

	resp, err := s.client.PostWithRetry(ctx, url, headers, body)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// prepareRequestBody: json → gzip → RSA (если задан публичный ключ).
func (s *MetricsSender) prepareRequestBody(payload []byte) ([]byte, error) {
	gz, err := s.compressGzip(payload)
	if err != nil {
		return nil, err
	}
	if s.publicKey == nil {
		return gz.Bytes(), nil
	}
	return crypto.Encrypt(s.publicKey, gz.Bytes())
}

func (s *MetricsSender) compressGzip(payload []byte) (bytes.Buffer, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(payload); err != nil {
		_ = gz.Close()
		return bytes.Buffer{}, fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return bytes.Buffer{}, fmt.Errorf("gzip close: %w", err)
	}
	return buf, nil
}

func (s *MetricsSender) hostIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			return ip4.String()
		}
	}
	return ""
}
