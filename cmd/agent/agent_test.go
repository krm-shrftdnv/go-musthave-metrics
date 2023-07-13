package main

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

var gaugeTestMetrics = map[string]*internal.Metric[internal.Gauge]{}

func Test_updateMetric(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				name: "Alloc",
			},
		},
	}
	for _, metricName := range metricNames {
		gaugeTestMetrics[metricName] = &internal.Metric[internal.Gauge]{
			Name: metricName,
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldValue := &gaugeTestMetrics[tt.args.name].Value
			updateMetric(tt.args.name)
			assert.Equal(t, oldValue, &gaugeTestMetrics[tt.args.name].Value, "updateMetric() = %v, want %v", oldValue, &gaugeTestMetrics[tt.args.name].Value)
		})
	}
}
