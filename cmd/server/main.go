package main

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"log"
	"os"
)

func main() {

	address, err := parseServerFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: address,
		},
	}

	app := internal.NewApp(cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
