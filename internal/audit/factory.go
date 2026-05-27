package audit

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"log"
)

// NewFromConfig собирает Publisher из ServerConfig (AuditFile, AuditURL).
func NewFromConfig(cfg config.ServerConfig) *Publisher {
	var observers []Observer
	if cfg.AuditFile != "" {
		sink, err := NewFileSink(cfg.AuditFile)
		if err != nil {
			log.Printf("audit file sink open %q: %v", cfg.AuditFile, err)
		} else {
			observers = append(observers, sink)
		}
	}
	if cfg.AuditURL != "" {
		observers = append(observers, NewHTTPSink(cfg.AuditURL))
	}
	return NewPublisher(observers...)
}
