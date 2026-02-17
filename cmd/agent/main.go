package main

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/agent"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/environmentVar"
	"log"
	"os"
)

func main() {
	envVar, err := environmentVar.LoadEnvVar()
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

	cfg := &config.Config{Client: client}

	agt := agent.NewAgent(cfg)

	agt.Start()

	select {}
}
