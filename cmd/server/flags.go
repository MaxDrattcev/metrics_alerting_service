package main

import (
	"flag"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/flags"
)

type ServerFlags struct {
	Address         string
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	Key             string
	AuditFile       string
	AuditURL        string
	CryptoKey       string
	Config          string
	TrustedSubnet   string
	GRPCAddress     string
	GRPCCert        string
	GRPCKey         string
}

func parseServerFlags() (*ServerFlags, error) {
	var (
		address         = flag.String("a", "", "адрес и порт сервера")
		storeInterval   = flag.Int("i", 0, "периодичность сохранения метрик в файл")
		fileStoragePath = flag.String("f", "", "путь до файла")
		restore         = flag.Bool("r", false, "загружать данные из файла при старте сервера")
		dataBaseDSN     = flag.String("d", "", "строка адреса подключения")
		key             = flag.String("k", "", "Ключ")
		auditFile       = flag.String("audit-file", "", "путь к файлу с логами аудита")
		auditURL        = flag.String("audit-url", "", "полный url по которому отправляются логи аудита")
		cryptoKey       = flag.String("crypto-key", "", "путь к файлу с приватным ключом")
		trustedSubnet   = flag.String("t", "", "строковое представление бесклассовой адресации (CIDR)")
		grpcAddress     = flag.String("g", "", "gRPC адрес сервера")
		grpcCert        = flag.String("grpc-cert", "", "путь к TLS-сертификату gRPC")
		grpcKey         = flag.String("grpc-key", "", "путь к приватному ключу gRPC")

		config string
	)
	flag.StringVar(&config, "config", "config.json", "имя файла конфигурации")
	flag.StringVar(&config, "c", "config.json", "имя файла конфигурации")

	flag.Parse()

	if err := flags.CheckUnknownFlags(); err != nil {
		return nil, err
	}

	return &ServerFlags{
		Address:         *address,
		StoreInterval:   int64(*storeInterval),
		FileStoragePath: *fileStoragePath,
		Restore:         *restore,
		DatabaseDSN:     *dataBaseDSN,
		Key:             *key,
		AuditFile:       *auditFile,
		AuditURL:        *auditURL,
		CryptoKey:       *cryptoKey,
		TrustedSubnet:   *trustedSubnet,
		Config:          config,
		GRPCAddress:     *grpcAddress,
		GRPCCert:        *grpcCert,
		GRPCKey:         *grpcKey,
	}, nil
}
