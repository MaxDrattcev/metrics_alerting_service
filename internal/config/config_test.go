package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClientConfig_GetPollInterval(t *testing.T) {
	tests := []struct {
		name         string
		pollInterval int64
		want         time.Duration
	}{
		{
			name:         "2 seconds",
			pollInterval: 2,
			want:         2 * time.Second,
		},
		{
			name:         "10 seconds",
			pollInterval: 10,
			want:         10 * time.Second,
		},
		{
			name:         "zero",
			pollInterval: 0,
			want:         0 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{
				PollInterval: tt.pollInterval,
			}

			interval := cfg.GetPollInterval()
			assert.Equal(t, tt.want, interval)
		})
	}
}

func TestClientConfig_GetReportInterval(t *testing.T) {
	tests := []struct {
		name           string
		reportInterval int64
		want           time.Duration
	}{
		{
			name:           "10 seconds",
			reportInterval: 10,
			want:           10 * time.Second,
		},
		{
			name:           "30 seconds",
			reportInterval: 30,
			want:           30 * time.Second,
		},
		{
			name:           "zero",
			reportInterval: 0,
			want:           0 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ClientConfig{
				ReportInterval: tt.reportInterval,
			}

			interval := cfg.GetReportInterval()
			assert.Equal(t, tt.want, interval)
		})
	}
}
