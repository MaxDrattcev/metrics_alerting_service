package main

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/agent"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentvar"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	envVar, err := environmentvar.LoadEnvVar()
	if err != nil {
		log.Printf("Warning: invalid environment variables, using flags/defaults: %v", err)
	}
	flags, err := parseAgentFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	var client = config.ClientConfig{
		Address:        envVar.Address,
		ReportInterval: envVar.ReportInterval,
		PollInterval:   envVar.PollInterval,
		Key:            envVar.Key,
		RateLimit:      envVar.RateLimit,
	}
	if client.Address == "" {
		client.Address = flags.Address
	}
	if client.ReportInterval == 0 {
		client.ReportInterval = flags.ReportInterval
	}
	if client.PollInterval == 0 {
		client.PollInterval = flags.PollInterval
	}
	if client.Key == "" {
		client.Key = flags.Key
	}
	if client.RateLimit == 0 {
		client.RateLimit = flags.RateLimit
	}
	cfg := &config.Config{Client: client}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	agt := agent.NewAgent(cfg)
	agt.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down agent...")
	cancel()
}
