package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testCounter Counter = 1
var testGauge Gauge = 2.2

func TestCounter_GetTypeName(t *testing.T) {
	counterTests := []struct {
		name string
		c    Counter
		want MetricTypeName
	}{
		{
			name: "success counter",
			c:    testCounter,
			want: CounterName,
		},
	}
	for _, tt := range counterTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.GetTypeName(), "GetTypeName() = %v, want %v", tt.want, tt.c.GetTypeName())
		})
	}
}

func TestGauge_GetTypeName(t *testing.T) {
	gaugeTests := []struct {
		name string
		g    Gauge
		want MetricTypeName
	}{
		{
			name: "success gauge",
			g:    testGauge,
			want: GaugeName,
		},
	}
	for _, tt := range gaugeTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.g.GetTypeName(), "GetTypeName() = %v, want %v", tt.want, tt.g.GetTypeName())
		})
	}
}
