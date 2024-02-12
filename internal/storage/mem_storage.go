package storage

import (
	"fmt"
	"strings"
	"sync"
)

var lock = sync.RWMutex{}

type MemStorage[T Element] struct {
	storage map[string]*T
}

func (ms *MemStorage[T]) Set(key string, value T) {
	lock.Lock()
	defer lock.Unlock()
	if ms.storage == nil {
		ms.Init()
	}
	ms.storage[key] = &value
}

func (ms *MemStorage[T]) Get(key string) (*T, bool) {
	lock.RLock()
	defer lock.RUnlock()
	value, ok := ms.storage[key]
	return value, ok
}

func (ms *MemStorage[T]) GetAll() map[string]*T {
	lock.RLock()
	defer lock.RUnlock()
	return ms.storage
}

func (ms *MemStorage[T]) Init() {
	if ms.storage == nil {
		ms.storage = make(map[string]*T)
	}
}

func (ms *MemStorage[Element]) String() string {
	lock.RLock()
	defer lock.RUnlock()
	sb := &strings.Builder{}
	for k, v := range ms.storage {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprint(v))
		sb.WriteString(",\n")
	}
	return sb.String()
}
