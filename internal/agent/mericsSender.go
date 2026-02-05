package agent

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

const (
	contentType = "Content-Type"
	textType    = "text/plain"
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
