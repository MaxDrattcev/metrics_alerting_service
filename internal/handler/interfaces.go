package handler

import "net/http"

type MetricsHandler interface {
	Update(http.ResponseWriter, *http.Request)
}
