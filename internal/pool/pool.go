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
	new  func() T
}

// New создаёт пул объектов типа T.
// newFunc вызывается, когда в пуле нет свободного объекта.
func New[T Resetter](newFunc func() T) *Pool[T] {
	p := &Pool[T]{new: newFunc}
	p.pool = sync.Pool{
		New: func() any {
			return newFunc()
		},
	}
	return p
}

// Get возвращает объект из пула или новый через newFunc.
func (p *Pool[T]) Get() T {
	v := p.pool.Get()
	if v == nil {
		return p.new()
	}
	return v.(T)
}

// Put сбрасывает состояние объекта и возвращает его в пул.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
