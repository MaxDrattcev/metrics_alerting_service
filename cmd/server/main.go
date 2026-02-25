package main

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"github.com/bytedance/gopkg/util/logger"
	"log"
	"os"
)

func main() {

	envVar, err := environmentvar.LoadEnvVar()
	if err != nil {
		logger.Info(err)
	}

	flags, err := parseServerFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	var server = config.ServerConfig{
		Address:         envVar.Address,
		StoreInterval:   envVar.StoreInterval,
		FileStoragePath: envVar.FileStoragePath,
		Restore:         envVar.Restore,
	}
	if server.Address == "" {
		server.Address = flags.Address
	}
	if server.StoreInterval == nil {
		server.StoreInterval = &flags.StoreInterval
	}
	if server.FileStoragePath == "" {
		server.FileStoragePath = flags.FileStoragePath
	}
	if server.Restore == nil {
		server.Restore = &flags.Restore
	}

	cfg := &config.Config{Server: server}

	app := internal.NewApp(cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
