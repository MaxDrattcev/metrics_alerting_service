package audit

// Observer — наблюдатель (приёмник события аудита).
type Observer interface {
	// Notify отправляет событие в приёмник (файл или HTTP).
	Notify(event Event) error
}
