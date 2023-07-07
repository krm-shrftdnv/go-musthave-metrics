package internal

type MetricTypeName string

const (
	GaugeName   MetricTypeName = "gauge"
	CounterName MetricTypeName = "counter"
)

type Gauge float64
type Counter int64
type MetricType interface {
	Gauge | Counter
	GetTypeName() MetricTypeName
}

type Metric[T MetricType] struct {
	Name  string
	Value T
}

func (g Gauge) GetTypeName() MetricTypeName {
	return GaugeName
}

func (c Counter) GetTypeName() MetricTypeName {
	return CounterName
}
