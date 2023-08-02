package storage

import (
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	Alloc         = internal.Gauge(1.1)
	BuckHashSys   = internal.Gauge(2.2)
	Frees         = internal.Gauge(3.3)
	GCCPUFraction = internal.Gauge(4.4)
	GCSys         = internal.Gauge(5.5)
	HeapAlloc     = internal.Gauge(6.6)
	HeapIdle      = internal.Gauge(7.7)
	HeapInuse     = internal.Gauge(8.8)
	HeapObjects   = internal.Gauge(9.9)
	HeapReleased  = internal.Gauge(0)
)

var testStorage = MemStorage[internal.Gauge]{
	storage: map[string]*internal.Gauge{
		"Alloc":         &Alloc,
		"BuckHashSys":   &BuckHashSys,
		"Frees":         &Frees,
		"GCCPUFraction": &GCCPUFraction,
		"GCSys":         &GCSys,
		"HeapAlloc":     &HeapAlloc,
		"HeapIdle":      &HeapIdle,
		"HeapInuse":     &HeapInuse,
		"HeapObjects":   &HeapObjects,
		"HeapReleased":  &HeapReleased,
	},
}
var emptyStorage = MemStorage[internal.Gauge]{}

func TestMemStorage_Get(t *testing.T) {
	type testCase[T Element] struct {
		name    string
		ms      *MemStorage[T]
		key     string
		want    *T
		wantErr bool
	}
	tests := []testCase[internal.Gauge]{
		{
			name:    "success",
			ms:      &testStorage,
			key:     "Alloc",
			want:    &Alloc,
			wantErr: false,
		},
		{
			name:    "element not found",
			ms:      &testStorage,
			key:     "BuckHashSys1",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element, ok := tt.ms.Get(tt.key)
			if !ok {
				if !tt.wantErr {
					t.Error("Get() = element not found")
				} else {
					assert.Equal(t, tt.want, element, "Get() = %v, want %v", tt.want, element)
				}
			}
			assert.Equal(t, tt.want, element, "Get() = %v, want %v", tt.want, element)
		})
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	type testCase[T Element] struct {
		name string
		ms   *MemStorage[T]
		want map[string]*T
	}
	tests := []testCase[internal.Gauge]{
		{
			name: "success",
			ms:   &testStorage,
			want: map[string]*internal.Gauge{
				"Alloc":         &Alloc,
				"BuckHashSys":   &BuckHashSys,
				"Frees":         &Frees,
				"GCCPUFraction": &GCCPUFraction,
				"GCSys":         &GCSys,
				"HeapAlloc":     &HeapAlloc,
				"HeapIdle":      &HeapIdle,
				"HeapInuse":     &HeapInuse,
				"HeapObjects":   &HeapObjects,
				"HeapReleased":  &HeapReleased,
			},
		},
		{
			name: "empty",
			ms:   &emptyStorage,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ms.GetAll(), "GetAll() = %v, want %v", tt.want, tt.ms.GetAll())
			if got := tt.ms.GetAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_Init(t *testing.T) {
	type testCase[T Element] struct {
		name string
		ms   *MemStorage[T]
	}
	tests := []testCase[internal.Gauge]{
		{
			name: "success init",
			ms:   &emptyStorage,
		},
		{
			name: "success no init",
			ms:   &testStorage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ms.Init()
			assert.NotNil(t, tt.ms.storage, "Init() = %v, failed", tt.ms.storage)
		})
	}
}

func TestMemStorage_Set(t *testing.T) {
	type args[T Element] struct {
		key   string
		value *T
	}
	type testCase[T Element] struct {
		name string
		ms   *MemStorage[T]
		args args[T]
		want args[T]
	}
	tests := []testCase[internal.Gauge]{
		{
			name: "success",
			ms:   &testStorage,
			args: args[internal.Gauge]{
				key:   "Alloc",
				value: &Alloc,
			},
			want: args[internal.Gauge]{
				key:   "Alloc",
				value: &Alloc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ms.Set(tt.args.key, *tt.args.value)
			element, ok := tt.ms.Get(tt.args.key)
			if !ok {
				t.Error("Get() = element not found")
			}
			assert.Equal(t, tt.want.value, element, "Set() = %v, want %v", tt.want, element)
		})
	}
}
