package storage

type FileStorage[T Element] struct {
	*MemStorage[T]
	FilePath string
}
