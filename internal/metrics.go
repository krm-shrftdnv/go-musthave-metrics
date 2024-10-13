package internal

import (
	"database/sql/driver"
	"strconv"
)

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
	String() string
}

type Metric[T MetricType] struct {
	Name  string
	Value T
}

func (g Gauge) GetTypeName() MetricTypeName {
	return GaugeName
}

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (c Counter) GetTypeName() MetricTypeName {
	return CounterName
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func (c Counter) Value() (driver.Value, error) {
	return int64(c), nil
}

func (c *Counter) Scan(value interface{}) error {
	if value == nil {
		*c = 0
		return nil
	}
	*c = Counter(value.(int64))
	return nil
}
