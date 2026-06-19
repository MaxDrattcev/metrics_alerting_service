package audit

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockObserver struct {
	events []Event
	err    error
	closed bool
}

func (m *mockObserver) Notify(event Event) error {
	m.events = append(m.events, event)
	return m.err
}

func (m *mockObserver) Close() error {
	m.closed = true
	return nil
}

func TestNewPublisher_EmptyObservers(t *testing.T) {
	require.Nil(t, NewPublisher())
}

func TestPublisher_PublishAndClose(t *testing.T) {
	obs := &mockObserver{}
	p := NewPublisher(obs)
	require.NotNil(t, p)

	event := Event{TS: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"}
	p.Publish(event)

	require.Eventually(t, func() bool {
		return len(obs.events) == 1
	}, time.Second, 10*time.Millisecond)

	require.NoError(t, p.Close())
	require.True(t, obs.closed)
}

func TestPublisher_PublishNil(t *testing.T) {
	var p *Publisher
	p.Publish(Event{})
	require.NoError(t, p.Close())
}

func TestPublisher_NotifyError(t *testing.T) {
	obs := &mockObserver{err: errors.New("notify failed")}
	p := NewPublisher(obs)
	p.Publish(Event{TS: 2})

	require.Eventually(t, func() bool {
		return len(obs.events) == 1
	}, time.Second, 10*time.Millisecond)

	require.NoError(t, p.Close())
}

func TestPublisher_CloseTwice(t *testing.T) {
	obs := &mockObserver{}
	p := NewPublisher(obs)
	require.NoError(t, p.Close())
	require.NoError(t, p.Close())
}

var _ io.Closer = (*mockObserver)(nil)

func TestPublisher_QueueFull(t *testing.T) {
	slow := &slowObserver{block: make(chan struct{})}
	p := NewPublisher(slow)
	for i := 0; i < auditQueueSize+10; i++ {
		p.Publish(Event{TS: int64(i)})
	}
	close(slow.block) // сначала разблокировать воркеров
	require.NoError(t, p.Close())
}

type slowObserver struct {
	block chan struct{}
}

func (s *slowObserver) Notify(event Event) error {
	<-s.block
	return nil
}
