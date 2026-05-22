package audit

import (
	"encoding/json"
	"os"
)

// FileSink записывает события аудита в файл (append, JSON-строка на строку).
type FileSink struct {
	path string
}

// NewFileSink создаёт файловый приёмник аудита.
func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

func (s *FileSink) Notify(event Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}
