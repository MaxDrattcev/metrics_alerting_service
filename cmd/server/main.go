package main

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config/db"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"github.com/bytedance/gopkg/util/logger"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"time"
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
		DatabaseDSN:     envVar.DatabaseDSN,
		Key:             envVar.Key,
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
	if server.DatabaseDSN == "" {
		server.DatabaseDSN = flags.DatabaseDSN
	}
	if server.Key == "" {
		server.Key = flags.Key
	}
	cfg := &config.Config{Server: server}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewConDB(ctx, *cfg, "file://migrations")
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	if pool != nil {
		defer pool.Close()
	}
	app := internal.NewApp(cfg, pool)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
