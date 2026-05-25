package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPSink отправляет события аудита POST-запросом в JSON.
type HTTPSink struct {
	url    string
	client httpDoer
}

// NewHTTPSink создаёт HTTP-приёмник аудита с автоматическими ретраями.
func NewHTTPSink(url string) *HTTPSink {
	return &HTTPSink{
		url:    url,
		client: newRetryHTTPClient(5 * time.Second),
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("audit post status: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
