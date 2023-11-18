package storage

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
)

type Element interface {
	internal.MetricType
	String() string
}

type Storage[T Element] interface {
	Set(key string, value T)
	Get(key string) (*T, bool)
	GetAll() map[string]*T
	String() string
}
