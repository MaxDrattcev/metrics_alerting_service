package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileSink записывает события аудита в файл (append, JSON-строка на строку).
type FileSink struct {
	path string
	f    *os.File
	mu   sync.Mutex
}

// NewFileSink создаёт файловый приёмник и открывает файл для записи.
// Файл открыт на всё время жизни приёмника.
func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileSink{
		path: path,
		f:    f,
	}, nil
}

func (s *FileSink) Notify(event Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f == nil {
		return os.ErrClosed
	}
	_, err = s.f.Write(append(b, '\n'))
	return err
}

func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f == nil {
		return nil
	}
	err := s.f.Close()
	s.f = nil
	return err
}
