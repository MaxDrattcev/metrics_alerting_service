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
}

func parseServerFlags() (*ServerFlags, error) {
	var (
		address         = flag.String("a", "localhost:8080", "адрес и порт сервера")
		storeInterval   = flag.Int("i", 300, "периодичность сохранения метрик в файл")
		fileStoragePath = flag.String("f", "metrics.json", "путь до файла")
		restore         = flag.Bool("r", false, "загружать данные из файла при старте сервера")
		dataBaseDSN     = flag.String("d", "", "строка адреса подключения")
		key             = flag.String("k", "", "Ключ")
		auditFile       = flag.String("audit-file", "", "путь к файлу с логами аудита")
		auditURL        = flag.String("audit-url", "", "полный url по которому отправляются логи аудита")
	)

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
	}, nil
}
