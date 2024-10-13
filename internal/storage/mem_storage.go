package storage

import (
	"fmt"
	"strings"
	"sync"
)

type MemStorage[T Element] struct {
	mx      sync.RWMutex
	storage map[string]*T
}

func (ms *MemStorage[T]) Set(key string, value T) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if ms.storage == nil {
		ms.mx.Unlock()
		ms.Init()
		ms.mx.Lock()
	}
	ms.storage[key] = &value
}

func (ms *MemStorage[T]) Get(key string) (*T, bool) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	value, ok := ms.storage[key]
	return value, ok
}

func (ms *MemStorage[T]) GetAll() map[string]*T {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	return ms.storage
}

func (ms *MemStorage[T]) Init() {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	if ms.storage == nil {
		ms.storage = make(map[string]*T)
	}
}

func (ms *MemStorage[Element]) String() string {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	sb := &strings.Builder{}
	for k, v := range ms.storage {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprint(v))
		sb.WriteString(",\n")
	}
	return sb.String()
}
