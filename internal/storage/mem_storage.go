package storage

import (
	"fmt"
	"strings"
)

type MemStorage[T Element] struct {
	storage map[string]*T
}

func (ms *MemStorage[T]) Set(key string, value T) {
	if ms.storage == nil {
		ms.Init()
	}
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

func (ms *MemStorage[Element]) String() string {
	sb := &strings.Builder{}
	for k, v := range ms.storage {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprint(v))
		sb.WriteString(",\n")
	}
	return sb.String()
}
