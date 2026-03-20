package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id" db:"id"`
	MType string   `json:"type" db:"type"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"hash,omitempty" db:"hash"`
}
