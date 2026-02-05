package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type AgentFlags struct {
	Address        string
	ReportInterval int64
	PollInterval   int64
}

func parseAgentFlags() (*AgentFlags, error) {
	var (
		address        = flag.String("a", "localhost:8080", "адрес и порт запуска HTTP-сервера")
		reportInterval = flag.Int("r", 10, "частота отправки метрик на сервер (в секундах)")
		pollInterval   = flag.Int("p", 2, "частота опроса метрик из пакета runtime (в секундах)")
	)

	flag.Parse()

	if err := checkUnknownFlags(); err != nil {
		return nil, err
	}

	return &AgentFlags{
		Address:        *address,
		ReportInterval: int64(*reportInterval),
		PollInterval:   int64(*pollInterval),
	}, nil
}

func checkUnknownFlags() error {
	knownFlags := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		knownFlags[f.Name] = true
	})

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if !strings.HasPrefix(arg, "-") {
			continue
		}

		flagName := strings.TrimPrefix(arg, "-")
		flagName = strings.TrimPrefix(flagName, "-")

		if idx := strings.Index(flagName, "="); idx != -1 {
			flagName = flagName[:idx]
		}

		if !knownFlags[flagName] {
			return fmt.Errorf("unknown flag: -%s", flagName)
		}
	}

	return nil
}
