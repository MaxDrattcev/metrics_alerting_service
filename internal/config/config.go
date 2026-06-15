package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config — корневая конфигурация приложения.
type Config struct {
	Server ServerConfig `json:"server"`
	Client ClientConfig `json:"client"`
}

// ServerConfig — параметры HTTP-сервера сбора метрик.
type ServerConfig struct {
	Address         string `json:"address"`
	StoreInterval   *int64 `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	Restore         *bool  `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	Key             string `json:"key"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	CryptoKey       string `json:"crypto_key"`
	ConfigServer    string
}

// ClientConfig — параметры агента (адрес сервера, интервалы, ключ).
type ClientConfig struct {
	Address        string `json:"address"`
	PollInterval   int64  `json:"poll_interval"`
	ReportInterval int64  `json:"report_interval"`
	Key            string `json:"key"`
	RateLimit      int    `json:"rate_limit"`
	CryptoKey      string `json:"crypto_key"`
	ConfigAgent    string
}

// GetStoreInterval возвращает интервал сохранения метрик в файл.
func (s *ServerConfig) GetStoreInterval() time.Duration {
	return time.Duration(*s.StoreInterval) * time.Second
}

// GetPollInterval возвращает интервал опроса метрик агентом.
func (c *ClientConfig) GetPollInterval() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}

// GetReportInterval возвращает интервал отправки метрик на сервер.
func (c *ClientConfig) GetReportInterval() time.Duration {
	return time.Duration(c.ReportInterval) * time.Second
}

// LoadConfigJSON читает JSON файл конфигурации и инициализирует структуру с ConfigJSON конфигурацией
func LoadConfigJSON(configPath string) (*Config, error) {

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}
