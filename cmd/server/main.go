package main

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config/db"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
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
		DataBaseDSN:     envVar.DataBaseDSN,
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
	if server.DataBaseDSN == "" {
		server.DataBaseDSN = flags.DataBaseDSN
	}

	cfg := &config.Config{Server: server}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pool *pgxpool.Pool
	if cfg.Server.DataBaseDSN != "" {
		var err error
		pool, err = db.NewPool(ctx, cfg.Server.DataBaseDSN)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer pool.Close()

		m, err := migrate.New("file://migrations", cfg.Server.DataBaseDSN)
		if err != nil {
			log.Fatalf("migrate init: %v", err)
		}
		defer m.Close()

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
	}
	app := internal.NewApp(cfg, pool)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
