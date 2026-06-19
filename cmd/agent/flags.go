package main

import (
	"flag"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/flags"
)

type AgentFlags struct {
	Address        string
	ReportInterval int64
	PollInterval   int64
	Key            string
	RateLimit      int
	CryptoKey      string
	Config         string
	GRPCAddress    string
}

func parseAgentFlags() (*AgentFlags, error) {
	var (
		address        = flag.String("a", "", "адрес и порт запуска HTTP-сервера")
		reportInterval = flag.Int("r", 0, "частота отправки метрик на сервер (в секундах)")
		pollInterval   = flag.Int("p", 0, "частота опроса метрик из пакета runtime (в секундах)")
		key            = flag.String("k", "", "Ключ")
		rateLimit      = flag.Int("l", 0, "количество одновременно исходящих запросов")
		cryptoKey      = flag.String("crypto-key", "", "путь к файлу с публичным ключом")
		grpcAddress    = flag.String("g", "", "gRPC адрес агента")
		config         string
	)
	flag.StringVar(&config, "config", "config.json", "имя файла конфигурации")
	flag.StringVar(&config, "c", "config.json", "имя файла конфигурации")

	flag.Parse()

	if err := flags.CheckUnknownFlags(); err != nil {
		return nil, err
	}

	return &AgentFlags{
		Address:        *address,
		ReportInterval: int64(*reportInterval),
		PollInterval:   int64(*pollInterval),
		Key:            *key,
		RateLimit:      *rateLimit,
		CryptoKey:      *cryptoKey,
		Config:         config,
		GRPCAddress:    *grpcAddress,
	}, nil
}
