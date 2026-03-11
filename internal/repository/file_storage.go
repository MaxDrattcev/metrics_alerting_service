package repository

import (
	"bufio"
	"encoding/json"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"os"
	"sync"
)

type fileStorage struct {
	path string
	mu   sync.Mutex
}

func NewFileStorage(path string) FileStorage {
	return &fileStorage{
		path: path,
	}
}

func (f *fileStorage) WriteMetrics(metrics []models.Metrics) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	writeFile, err := os.OpenFile(f.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer writeFile.Close()

	writer := bufio.NewWriter(writeFile)

	data, err := json.Marshal(&metrics)
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return err
	}

	return writer.Flush()
}

func (f *fileStorage) ReadMetrics() ([]models.Metrics, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	readFile, err := os.OpenFile(f.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return []models.Metrics{}, err
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	data := scanner.Bytes()

	metrics := []models.Metrics{}
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
