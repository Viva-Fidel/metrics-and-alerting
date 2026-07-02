package pool

import (
	"reflect"
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	pool sync.Pool
}

func New[T Resetter](new func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return new()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	if isNil(x) {
		return
	}
	x.Reset()
	p.pool.Put(x)
}

func isNil[T any](v T) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
