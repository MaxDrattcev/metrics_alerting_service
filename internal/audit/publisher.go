package audit

import "log"

// Publisher — субъект: рассылает событие всем подписчикам.
type Publisher struct {
	observers []Observer
}

// NewPublisher создаёт издателя; при пустом списке observers возвращает nil.
func NewPublisher(observers ...Observer) *Publisher {
	if len(observers) == 0 {
		return nil
	}
	return &Publisher{observers: observers}
}

// Publish уведомляет всех наблюдателей; ошибки логируются, не прерывают запрос.
func (p *Publisher) Publish(event Event) {
	if p == nil {
		return
	}
	for _, o := range p.observers {
		if err := o.Notify(event); err != nil {
			log.Printf("audit notify error: %v", err)
		}
	}
}
