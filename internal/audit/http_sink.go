package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPSink отправляет события аудита POST-запросом в JSON.
type HTTPSink struct {
	url    string
	client *http.Client
}

// NewHTTPSink создаёт HTTP-приёмник аудита.
func NewHTTPSink(url string) *HTTPSink {
	return &HTTPSink{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *HTTPSink) Notify(event Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit post status: %d", resp.StatusCode)
	}
	return nil
}
