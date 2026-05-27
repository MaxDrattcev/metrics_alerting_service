package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config — корневая конфигурация приложения.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Client ClientConfig `yaml:"client"`
}

// ServerConfig — параметры HTTP-сервера сбора метрик.
type ServerConfig struct {
	Address         string `yaml:"address"`
	StoreInterval   *int64 `yaml:"storeInterval"`
	FileStoragePath string `yaml:"fileStoragePath"`
	Restore         *bool  `yaml:"restore"`
	DatabaseDSN     string `yaml:"databaseDSN"`
	Key             string `yaml:"key"`
	AuditFile       string `yaml:"auditFile"`
	AuditURL        string `yaml:"auditUrl"`
}

// ClientConfig — параметры агента (адрес сервера, интервалы, ключ).
type ClientConfig struct {
	Address        string `yaml:"address"`
	PollInterval   int64  `yaml:"pollInterval"`
	ReportInterval int64  `yaml:"reportInterval"`
	Key            string `yaml:"key"`
	RateLimit      int    `yaml:"rateLimit"`
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

func Load() (*Config, error) {
	configPath := "config/config.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}
