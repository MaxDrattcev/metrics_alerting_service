package service

type MetricsService interface {
	UpdateGauge(string, string, *float64) error

	UpdateCounter(string, string, *int64) error
}
