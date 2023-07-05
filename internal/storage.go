package internal

type Element interface {
	int64 | float64
}

type Storage[T Element] interface {
	Set(key string, value T)
	Get(key string) T
	GetAll() map[string]T
}

type MemStorage[T Element] struct {
	storage map[string]T
}

func (ms *MemStorage[Element]) Set(key string, value Element) {
	ms.storage[key] = value
}

func (ms *MemStorage[Element]) Get(key string) Element {
	return ms.storage[key]
}

func (ms *MemStorage[Element]) GetAll() map[string]Element {
	return ms.storage
}

func (ms *MemStorage[Element]) Init() {
	if ms.storage == nil {
		ms.storage = make(map[string]Element)
	}
}
