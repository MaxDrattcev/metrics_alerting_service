package environmentvar

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

// EnvVar — переменные окружения сервера и агента.
type EnvVar struct {
	Address         string `env:"ADDRESS"`
	ReportInterval  int64  `env:"REPORT_INTERVAL"`
	PollInterval    int64  `env:"POLL_INTERVAL"`
	StoreInterval   *int64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	RateLimit       int    `env:"RATE_LIMIT"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	CryptoKeyAgent  string `env:"CRYPTO_KEY_AGENT"`
	CryptoKeyServer string `env:"CRYPTO_KEY_SERVER"`
	ConfigAgent     string `env:"CONFIG_AGENT"`
	ConfigServer    string `env:"CONFIG_SERVER"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	GRPCAddress     string `env:"GRPC_ADDRESS"`
}

// LoadEnvVar парсит переменные окружения в структуру EnvVar.
func LoadEnvVar() (EnvVar, error) {
	var envVar EnvVar
	err := env.Parse(&envVar)
	if err != nil {
		return EnvVar{}, fmt.Errorf("error parsing environment variables: %w", err)
	}
	return envVar, nil
}
