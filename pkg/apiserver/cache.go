package apiserver

import (
	"k8s.io/client-go/tools/cache"
)

type Store[T any] interface {

	// List returns a list of all the currently non-empty accumulators
	List() []T

	// ListKeys returns a list of all the keys currently associated with non-empty accumulators
	ListKeys() []string

	// GetByKey returns the accumulator associated with the given key
	GetByKey(key string) (item T, exists bool, err error)

	// Resync is meaningless in the terms appearing here but has
	// meaning in some implementations that have non-trivial
	// additional behavior (e.g., DeltaFIFO).
	Resync() error
}

func NewStore[T any](store cache.Store) Store[T] {
	return &storeWrapper[T]{
		store: store,
	}
}

type storeWrapper[T any] struct {
	store cache.Store
}

func (s *storeWrapper[T]) List() []T {
	objs := s.store.List()
	ret := make([]T, len(objs))
	for i, obj := range objs {
		ret[i] = obj.(T)
	}
	return ret
}

func (s *storeWrapper[T]) ListKeys() []string {
	return s.store.ListKeys()
}

func (s *storeWrapper[T]) GetByKey(key string) (item T, exists bool, err error) {
	obj, exists, err := s.store.GetByKey(key)
	if err != nil {
		return item, false, err
	}
	if !exists {
		return item, false, nil
	}
	return obj.(T), true, nil
}

func (s *storeWrapper[T]) Resync() error {
	return s.store.Resync()
}
