package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Client ClientConfig `yaml:"client"`
}

type ServerConfig struct {
	Address         string `yaml:"address"`
	StoreInterval   *int64 `yaml:"storeInterval"`
	FileStoragePath string `yaml:"fileStoragePath"`
	Restore         *bool  `yaml:"restore"`
	DatabaseDSN     string `yaml:"databaseDSN"`
}

type ClientConfig struct {
	Address        string `yaml:"address"`
	PollInterval   int64  `yaml:"pollInterval"`
	ReportInterval int64  `yaml:"reportInterval"`
}

func (s *ServerConfig) GetStoreInterval() time.Duration {
	return time.Duration(*s.StoreInterval) * time.Second
}

func (c *ClientConfig) GetPollInterval() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}

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
