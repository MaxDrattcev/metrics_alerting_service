package audit

import (
	"errors"
	"io"
	"log"
	"sync"
)

const (
	auditQueueSize  = 256
	auditMaxWorkers = 8
)

type publishJob struct {
	observer Observer
	event    Event
}

// Publisher — субъект: рассылает событие всем подписчикам.
type Publisher struct {
	observers []Observer
	jobs      chan publishJob
	wg        sync.WaitGroup
	stopOnce  sync.Once
}

// NewPublisher создаёт издателя; при пустом списке observers возвращает nil.
func NewPublisher(observers ...Observer) *Publisher {
	if len(observers) == 0 {
		return nil
	}

	workers := len(observers) * 2
	if workers < 2 {
		workers = 2
	}
	if workers > auditMaxWorkers {
		workers = auditMaxWorkers
	}

	p := &Publisher{
		observers: observers,
		jobs:      make(chan publishJob, auditQueueSize),
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p
}

func (p *Publisher) worker() {
	defer p.wg.Done()

	for job := range p.jobs {
		if err := job.observer.Notify(job.event); err != nil {
			log.Printf("audit notify error: %v", err)
		}
	}
}

// Publish ставит событие в очередь для каждого наблюдателя (не блокирует друг друга).
// При переполнении очереди событие для этого наблюдателя отбрасывается (логируется).
func (p *Publisher) Publish(event Event) {
	if p == nil {
		return
	}

	for _, o := range p.observers {
		job := publishJob{
			observer: o,
			event:    event,
		}

		select {
		case p.jobs <- job:
		default:
			log.Printf("audit: queue full, event dropped for %T", o)
		}
	}
}

// Close останавливает воркеров и закрывает ресурсы наблюдателей (файл аудита и т.д.).
func (p *Publisher) Close() error {
	if p == nil {
		return nil
	}

	var closeErr error

	p.stopOnce.Do(func() {
		close(p.jobs)
		p.wg.Wait()

		for _, o := range p.observers {
			c, ok := o.(io.Closer)
			if !ok {
				continue
			}
			if err := c.Close(); err != nil {
				closeErr = errors.Join(closeErr, err)
			}
		}
	})

	return closeErr
}
