package internal

type MetricTypeName string

const (
	gaugeName   MetricTypeName = "gauge"
	counterName MetricTypeName = "counter"
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
	return gaugeName
}

func (c Counter) GetTypeName() MetricTypeName {
	return counterName
}
