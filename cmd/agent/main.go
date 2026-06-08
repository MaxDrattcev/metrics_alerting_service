package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/buildinfo"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/agent"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
)

func main() {
	buildinfo.Print()

	envVar, err := environmentvar.LoadEnvVar()
	if err != nil {
		log.Printf("Warning: invalid environment variables, using flags/defaults: %v", err)
	}
	flags, err := parseAgentFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := initConfig(envVar, *flags)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	agt, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatalf("create agent: %v", err)
	}
	agt.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop

	log.Println("Shutting down agent...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	if err := agt.Shutdown(shutdownCtx); err != nil {
		log.Printf("agent shutdown: %v", err)
	}
}

func initConfig(envVar environmentvar.EnvVar, flags AgentFlags) (*config.Config, error) {
	var cfg *config.Config
	var err error
	configPath := envVar.ConfigAgent
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
		cfg.Client.Address = flags.Address
	}
	if flags.ReportInterval != 0 {
		cfg.Client.ReportInterval = flags.ReportInterval
	}
	if flags.PollInterval != 0 {
		cfg.Client.PollInterval = flags.PollInterval
	}
	if flags.Key != "" {
		cfg.Client.Key = flags.Key
	}
	if flags.RateLimit != 0 {
		cfg.Client.RateLimit = flags.RateLimit
	}
	if flags.CryptoKey != "" {
		cfg.Client.CryptoKey = flags.CryptoKey
	}

	if envVar.Address != "" {
		cfg.Client.Address = envVar.Address
	}
	if envVar.ReportInterval != 0 {
		cfg.Client.ReportInterval = envVar.ReportInterval
	}
	if envVar.PollInterval != 0 {
		cfg.Client.PollInterval = envVar.PollInterval
	}
	if envVar.Key != "" {
		cfg.Client.Key = envVar.Key
	}
	if envVar.RateLimit != 0 {
		cfg.Client.RateLimit = envVar.RateLimit
	}
	if envVar.CryptoKeyAgent != "" {
		cfg.Client.CryptoKey = envVar.CryptoKeyAgent
	}
	return cfg, nil
}
