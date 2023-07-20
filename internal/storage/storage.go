package storage

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"strings"
)

type Element interface {
	internal.MetricType
}

type Storage[T Element] interface {
	Set(key string, value T)
	Get(key string) (T, error)
	GetAll() map[string]T
}

type MemStorage[T Element] struct {
	storage map[string]T
}

func (ms *MemStorage[T]) Set(key string, value T) {
	ms.storage[key] = value
}

func (ms *MemStorage[T]) Get(key string) (T, bool) {
	value, ok := ms.storage[key]
	return value, ok
}

func (ms *MemStorage[T]) GetAll() map[string]T {
	return ms.storage
}

func (ms *MemStorage[T]) Init() {
	if ms.storage == nil {
		ms.storage = make(map[string]T)
	}
}

func (ms *MemStorage[T]) String() string {
	sb := &strings.Builder{}
	for k, v := range ms.storage {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v.String())
		sb.WriteString(",\n")
	}
	return sb.String()
}
