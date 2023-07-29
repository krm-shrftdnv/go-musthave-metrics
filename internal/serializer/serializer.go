package serializer

import "github.com/krm-shrftdnv/go-musthave-metrics/internal"

type Metrics struct {
	ID    string            `json:"id"`
	MType string            `json:"type"`
	Delta *internal.Counter `json:"delta,omitempty"`
	Value *internal.Gauge   `json:"value,omitempty"`
}
