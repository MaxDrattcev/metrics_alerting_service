package main

import (
	"flag"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/flags"
)

func parseServerFlags() (string, error) {
	var address = flag.String("a", "localhost:8080", "адрес и порт сервера")

	flag.Parse()

	if err := flags.CheckUnknownFlags(); err != nil {
		return "", err
	}

	return *address, nil
}
