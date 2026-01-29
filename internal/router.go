package internal

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"net/http"
)

func SetupRouter(metricsHandler handler.MetricsHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/", metricsHandler.Update)

	return mux
}
