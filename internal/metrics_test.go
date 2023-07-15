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

func TestCounter_String(t *testing.T) {
	tests := []struct {
		name string
		c    Counter
		want string
	}{
		{name: "success", c: testCounter, want: "1"},
		{name: "success zero", c: 0, want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.c.String(), "String()")
		})
	}
}

func TestGauge_String(t *testing.T) {
	tests := []struct {
		name string
		g    Gauge
		want string
	}{
		{name: "success", g: testGauge, want: "2.2"},
		{name: "success zero", g: 0, want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.g.String(), "String()")
		})
	}
}
