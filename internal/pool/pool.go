package pool

import "sync"

// Resetter is a constraint for types that have a Reset() method.
type Resetter interface {
	Reset()
}

// Pool is a generic wrapper around sync.Pool for objects of type T.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a new Pool. The newFn function is called to create
// a new object when the pool is empty.
func New[T Resetter](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get returns an object from the pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets the object's state and returns it to the pool.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
