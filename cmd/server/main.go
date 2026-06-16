package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/buildinfo"

	"github.com/MaxDrattcev/metrics_alerting_service/internal"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config/db"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"github.com/bytedance/gopkg/util/logger"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	buildinfo.Print()
	envVar, err := environmentvar.LoadEnvVar()
	if err != nil {
		logger.Info(err)
	}
	flags, err := parseServerFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cfg, err := initConfig(envVar, *flags)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := db.NewConDB(ctx, *cfg, "file://migrations")
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	app := internal.NewApp(cfg, pool)
	go func() {
		log.Println("pprof on http://localhost:6060/debug/pprof/")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("Failed to run app: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
}

func initConfig(envVar environmentvar.EnvVar, flags ServerFlags) (*config.Config, error) {
	var cfg *config.Config
	var err error
	configPath := envVar.ConfigServer
	if configPath == "" {
		configPath = flags.Config
	}
	if configPath != "" {
		cfg, err = config.LoadConfigJSON(configPath)
		if err != nil {
			return nil, err
		}
	} else {
		cfg = &config.Config{}
	}

	if flags.Address != "" {
		cfg.Server.Address = flags.Address
	}
	if flags.StoreInterval != 0 {
		cfg.Server.StoreInterval = &flags.StoreInterval
	}
	if flags.FileStoragePath != "" {
		cfg.Server.FileStoragePath = flags.FileStoragePath
	}
	if flags.Restore != false {
		cfg.Server.Restore = &flags.Restore
	}
	if flags.DatabaseDSN != "" {
		cfg.Server.DatabaseDSN = flags.DatabaseDSN
	}
	if flags.Key != "" {
		cfg.Server.Key = flags.Key
	}
	if flags.AuditFile != "" {
		cfg.Server.AuditFile = flags.AuditFile
	}
	if flags.AuditURL != "" {
		cfg.Server.AuditURL = flags.AuditURL
	}
	if flags.CryptoKey != "" {
		cfg.Server.CryptoKey = flags.CryptoKey
	}
	if flags.TrustedSubnet != "" {
		cfg.Server.TrustedSubnet = flags.TrustedSubnet
	}

	if envVar.Address != "" {
		cfg.Server.Address = envVar.Address
	}
	if envVar.StoreInterval != nil {
		cfg.Server.StoreInterval = envVar.StoreInterval
	}
	if envVar.FileStoragePath != "" {
		cfg.Server.FileStoragePath = envVar.FileStoragePath
	}
	if envVar.Restore != nil {
		cfg.Server.Restore = envVar.Restore
	}
	if envVar.DatabaseDSN != "" {
		cfg.Server.DatabaseDSN = envVar.DatabaseDSN
	}
	if envVar.Key != "" {
		cfg.Server.Key = envVar.Key
	}
	if envVar.AuditFile != "" {
		cfg.Server.AuditFile = envVar.AuditFile
	}
	if envVar.AuditURL != "" {
		cfg.Server.AuditURL = envVar.AuditURL
	}
	if envVar.CryptoKeyServer != "" {
		cfg.Server.CryptoKey = envVar.CryptoKeyServer
	}
	if envVar.TrustedSubnet != "" {
		cfg.Server.TrustedSubnet = envVar.TrustedSubnet
	}
	return cfg, nil
}
