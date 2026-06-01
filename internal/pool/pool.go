package pool

import "sync"

// Resetter — тип с методом сброса состояния (как у ResetableStruct).
type Resetter interface {
	Reset()
}

// Pool хранит и переиспользует объекты одного типа T.
// Перед возвратом в пул вызывается Reset().
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New создаёт пул объектов типа T.
// newFunc вызывается sync.Pool, когда в пуле нет свободного объекта.
func New[T Resetter](newFunc func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool = sync.Pool{
		New: func() any {
			return newFunc()
		},
	}
	return p
}

// Get возвращает объект из пула (при пустом пуле sync.Pool вызывает New).
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put сбрасывает состояние объекта и возвращает его в пул.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
