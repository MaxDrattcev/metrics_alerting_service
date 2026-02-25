package repository

import (
	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")

	storage := NewFileStorage(path)
	require.NotNil(t, storage)
}

func TestFileStorage_WriteMetrics_ReadMetrics_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")
	storage := NewFileStorage(path)

	metrics := []models.Metrics{
		{ID: "g1", MType: models.Gauge, Value: floatPtr(1.5)},
		{ID: "c1", MType: models.Counter, Delta: int64Ptr(10)},
	}

	err := storage.WriteMetrics(metrics)
	require.NoError(t, err)

	read, err := storage.ReadMetrics()
	require.NoError(t, err)
	require.Len(t, read, 2)
	require.Equal(t, "g1", read[0].ID)
	require.Equal(t, models.Gauge, read[0].MType)
	require.NotNil(t, read[0].Value)
	require.Equal(t, 1.5, *read[0].Value)
	require.Equal(t, "c1", read[1].ID)
	require.Equal(t, models.Counter, read[1].MType)
	require.NotNil(t, read[1].Delta)
	require.Equal(t, int64(10), *read[1].Delta)
}

func TestFileStorage_WriteMetrics_EmptySlice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	storage := NewFileStorage(path)

	err := storage.WriteMetrics([]models.Metrics{})
	require.NoError(t, err)

	read, err := storage.ReadMetrics()
	require.NoError(t, err)
	require.NotNil(t, read)
	require.Len(t, read, 0)
}

func TestFileStorage_ReadMetrics_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")
	storage := NewFileStorage(path)

	read, err := storage.ReadMetrics()
	require.NoError(t, err)
	require.Nil(t, read)
}

func TestFileStorage_WriteOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overwrite.json")
	storage := NewFileStorage(path)

	first := []models.Metrics{
		{ID: "old", MType: models.Gauge, Value: floatPtr(99.0)},
	}
	err := storage.WriteMetrics(first)
	require.NoError(t, err)

	second := []models.Metrics{
		{ID: "new", MType: models.Counter, Delta: int64Ptr(1)},
	}
	err = storage.WriteMetrics(second)
	require.NoError(t, err)

	read, err := storage.ReadMetrics()
	require.NoError(t, err)
	require.Len(t, read, 1)
	require.Equal(t, "new", read[0].ID)
	require.Equal(t, models.Counter, read[0].MType)
	require.NotNil(t, read[0].Delta)
	require.Equal(t, int64(1), *read[0].Delta)
}

func TestFileStorage_WriteMetrics_InvalidPath(t *testing.T) {
	path := filepath.Join("nonexistent", "dir", "metrics.json")
	storage := NewFileStorage(path)

	err := storage.WriteMetrics([]models.Metrics{
		{ID: "x", MType: models.Gauge, Value: floatPtr(1.0)},
	})
	require.Error(t, err)
}

func TestFileStorage_ReadMetrics_NoFile_CreatesEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile.json")
	storage := NewFileStorage(path)

	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err))

	read, err := storage.ReadMetrics()
	require.NoError(t, err)
	require.Nil(t, read)

	_, err = os.Stat(path)
	require.NoError(t, err)
}
