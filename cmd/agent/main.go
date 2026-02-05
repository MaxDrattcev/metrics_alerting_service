package main

import (
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/agent"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"os"
)

func main() {
	flags, err := parseAgentFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Config{
		Client: config.ClientConfig{
			Address:        flags.Address,
			ReportInterval: flags.ReportInterval,
			PollInterval:   flags.PollInterval,
		},
	}

	agt := agent.NewAgent(cfg)

	agt.Start()

	select {}
}
