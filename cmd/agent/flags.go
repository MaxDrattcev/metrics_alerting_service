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
}

func parseAgentFlags() (*AgentFlags, error) {
	var (
		address        = flag.String("a", "localhost:8080", "адрес и порт запуска HTTP-сервера")
		reportInterval = flag.Int("r", 10, "частота отправки метрик на сервер (в секундах)")
		pollInterval   = flag.Int("p", 2, "частота опроса метрик из пакета runtime (в секундах)")
		key            = flag.String("k", "", "Ключ")
		rateLimit      = flag.Int("l", 10, "количество одновременно исходящих запросов")
	)

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
	}, nil
}
