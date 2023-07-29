package storage

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
)

type Element interface {
	internal.MetricType
}

type Storage[T Element] interface {
	Set(key string, value T)
	Get(key string) (*T, bool)
	GetAll() map[string]*T
}

type MemStorage[T Element] struct {
	storage map[string]*T
}

func (ms *MemStorage[T]) Set(key string, value T) {
	ms.storage[key] = &value
}

func (ms *MemStorage[T]) Get(key string) (*T, bool) {
	value, ok := ms.storage[key]
	return value, ok
}

func (ms *MemStorage[T]) GetAll() map[string]*T {
	return ms.storage
}

func (ms *MemStorage[T]) Init() {
	if ms.storage == nil {
		ms.storage = make(map[string]*T)
	}
}
