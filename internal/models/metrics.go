package models

const (
	// Counter — тип метрики «счётчик».
	Counter = "counter"

	// Gauge — тип метрики «gauge» (измерение).
	Gauge = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.

// Metrics описывает одну метрику: имя, тип и значение.
// Value и Delta — указатели, чтобы отличать отсутствие значения от нуля.
type Metrics struct {
	// ID — уникальное имя метрики (например, Alloc, PollCount).
	ID string `json:"id" db:"id"`
	// MType — тип метрики: gauge или counter.
	MType string `json:"type" db:"type"`
	// Delta — значение counter: на сколько увеличить счётчик в этом запросе.
	Delta *int64 `json:"delta,omitempty" db:"delta"`
	// Value — значение для gauge.
	Value *float64 `json:"value,omitempty" db:"value"`
	// Hash — HMAC-подпись тела запроса
	Hash string `json:"hash,omitempty" db:"hash"`
}
