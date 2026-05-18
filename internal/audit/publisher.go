package audit

import "log"

type Publisher struct {
	observers []Observer
}

func NewPublisher(observers ...Observer) *Publisher {
	if len(observers) == 0 {
		return nil
	}
	return &Publisher{observers: observers}
}
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
