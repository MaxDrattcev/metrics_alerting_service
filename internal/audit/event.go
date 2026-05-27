package audit

// Event — событие аудита после успешного приёма метрик.
type Event struct {
	// TS — Unix timestamp события.
	TS int64 `json:"ts"`
	// Metrics — имена принятых метрик.
	Metrics []string `json:"metrics"`
	// IPAddress — IP входящего запроса.
	IPAddress string `json:"ip_address"`
}
