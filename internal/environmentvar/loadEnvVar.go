package environmentvar

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

type EnvVar struct {
	Address         string `env:"ADDRESS"`
	ReportInterval  int64  `env:"REPORT_INTERVAL"`
	PollInterval    int64  `env:"POLL_INTERVAL"`
	StoreInterval   *int64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
}

func LoadEnvVar() (EnvVar, error) {
	var envVar EnvVar
	err := env.Parse(&envVar)
	if err != nil {
		return EnvVar{}, fmt.Errorf("error parsing environment variables: %w", err)
	}
	return envVar, nil
}
