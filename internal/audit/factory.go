package audit

import "github.com/MaxDrattcev/metrics_alerting_service/internal/config"

// NewFromConfig собирает Publisher из ServerConfig (AuditFile, AuditURL).
func NewFromConfig(cfg config.ServerConfig) *Publisher {
	var observers []Observer
	if cfg.AuditFile != "" {
		observers = append(observers, NewFileSink(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		observers = append(observers, NewHTTPSink(cfg.AuditURL))
	}
	return NewPublisher(observers...)
}
