package config

import (
	"errors"
	"fmt"
	"sync"
)

// Factory is a stateful factory holding the necessary information to validate and construct T.
type Factory[T any] interface {
	New() (T, error)
	Validate() error
}

type unknownFactoryType[T any] struct {
	factoryType string
	typ         string
}

func (f unknownFactoryType[T]) New() (T, error) {
	var factory T

	return factory, fmt.Errorf("unknown %s type: %s", f.factoryType, f.typ)
}

func (f unknownFactoryType[T]) Validate() error {
	return fmt.Errorf("unknown %s type: %s", f.factoryType, f.typ)
}

// factoryRegistry holds a list of named factories of T.
type factoryRegistry[T any] struct {
	factories map[string]func() Factory[T]

	mu sync.RWMutex

	initOnce sync.Once
}

func (r *factoryRegistry[T]) init() {
	r.initOnce.Do(func() {
		if r.factories == nil {
			r.factories = make(map[string]func() Factory[T])
		}
	})
}

// RegisterFactory registers a factory of T.
//
// If RegisterFactory is called twice with the same name or if factory is nil, it returns an error.
func (r *factoryRegistry[T]) RegisterFactory(name string, factoryConstructor func() Factory[T]) error {
	r.init()

	r.mu.Lock()
	defer r.mu.Unlock()

	if factoryConstructor == nil {
		return errors.New("factory is nil")
	}

	if _, dup := r.factories[name]; dup {
		return errors.New("registration called twice for factory " + name)
	}

	r.factories[name] = factoryConstructor

	return nil
}

// GetFactory returns a factory.
//
// If a factory is not found, it returns false as the second argument.
func (r *factoryRegistry[T]) GetFactory(name string) (Factory[T], bool) {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()

	factoryConstructor, ok := r.factories[name]

	if !ok {
		var factory Factory[T]
		return factory, false
	}

	return factoryConstructor(), ok
}
