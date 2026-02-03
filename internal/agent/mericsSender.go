package agent

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"net/http"
	"time"
)

type MetricsSender struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewMetricsSender(cfg *config.Config) *MetricsSender {
	return &MetricsSender{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cfg: cfg,
	}
}

func (s *MetricsSender) SendGauge(name string, value float64) error {
	url := fmt.Sprintf("%s/update/gauge/%s/%v",
		s.cfg.Client.Address, name, value)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *MetricsSender) SendCounter(name string, value int64) error {
	url := fmt.Sprintf("%s/update/counter/%s/%d",
		s.cfg.Client.Address, name, value)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
