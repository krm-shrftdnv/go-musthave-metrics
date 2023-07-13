package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestStorageStateHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	type gaugeStateTestCase struct {
		name    string
		h       StorageStateHandler[internal.Gauge]
		want    want
		request string
	}
	type counterStateTestCase struct {
		name    string
		h       StorageStateHandler[internal.Counter]
		want    want
		request string
	}
	var gaugeStorage storage.MemStorage[internal.Gauge]
	var counterStorage storage.MemStorage[internal.Counter]
	gaugeStorage.Init()
	counterStorage.Init()

	gaugeStateHandler := StorageStateHandler[internal.Gauge]{
		Storage: &gaugeStorage,
	}
	counterStateHandler := StorageStateHandler[internal.Counter]{
		Storage: &counterStorage,
	}
	gaugeStateTests := []gaugeStateTestCase{
		{
			name: "success gauge",
			h:    gaugeStateHandler,
			want: want{
				code:        http.StatusOK,
				response:    "{}",
				contentType: "application/json",
			},
			request: "/state/gauge",
		},
	}
	counterStateTests := []counterStateTestCase{
		{
			name: "success counter",
			h:    counterStateHandler,
			want: want{
				code:        http.StatusOK,
				response:    "{}",
				contentType: "application/json",
			},
			request: "/state/counter",
		},
	}
	for _, tt := range gaugeStateTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(gaugeStateHandler.ServeHTTP)
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			gaugeStateResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			err = json.Unmarshal(gaugeStateResult, &gaugeStorage)
			require.NoError(t, err)

			require.NoError(t, err)
			assert.JSONEq(t, string(gaugeStateResult), tt.want.response)
		})
	}
	for _, tt := range counterStateTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(counterStateHandler.ServeHTTP)
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			counterStateResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			err = json.Unmarshal(counterStateResult, &counterStorage)
			require.NoError(t, err)

			require.NoError(t, err)
			assert.JSONEq(t, string(counterStateResult), tt.want.response)
		})
	}
}

func TestUpdateMetricHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code int
	}
	type updateMetricTestCase struct {
		name           string
		h              UpdateMetricHandler
		want           want
		method         string
		request        string
		metricTypeName string
		metricName     string
		metricValue    string
	}

	var gaugeStorage storage.MemStorage[internal.Gauge]
	var counterStorage storage.MemStorage[internal.Counter]
	gaugeStorage.Init()
	counterStorage.Init()

	updateMetricHandler := UpdateMetricHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	gaugeUpdateMetricTests := []updateMetricTestCase{
		{
			name: "success gauge update",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusOK,
			},
			method:         http.MethodPost,
			request:        "/update/%s/%s/%s",
			metricTypeName: string(internal.GaugeName),
			metricName:     "metric1",
			metricValue:    strconv.FormatFloat(123.45, 'f', -1, 64),
		},
		{
			name: "gauge wrong method",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusMethodNotAllowed,
			},
			method:         http.MethodGet,
			request:        "/update/%s/%s/%s",
			metricTypeName: string(internal.GaugeName),
			metricName:     "metric1",
			metricValue:    strconv.FormatFloat(123.45, 'f', -1, 64),
		},
		{
			name: "gauge wrong path",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusNotFound,
			},
			method:  http.MethodPost,
			request: fmt.Sprintf("/update/%s/%s", string(internal.GaugeName), strconv.FormatFloat(123.45, 'f', -1, 64)),
		},
		{
			name: "gauge wrong value",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusBadRequest,
			},
			method:         http.MethodPost,
			request:        "/update/%s/%s/%s",
			metricTypeName: string(internal.GaugeName),
			metricName:     "metric1",
			metricValue:    "value1",
		},
		{
			name: "gauge wrong type",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusBadRequest,
			},
			method:         http.MethodPost,
			request:        "/update/%s/%s/%s",
			metricTypeName: "unexpected",
			metricName:     "metric1",
			metricValue:    strconv.FormatFloat(123.45, 'f', -1, 64),
		},
	}
	counterMetricUpdateTests := []updateMetricTestCase{
		{
			name: "counter success update",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusOK,
			},
			method:         http.MethodPost,
			request:        "/update/%s/%s/%s",
			metricTypeName: string(internal.CounterName),
			metricName:     "metric2",
			metricValue:    strconv.FormatInt(123, 10),
		},
		{
			name: "counter wrong value",
			h:    updateMetricHandler,
			want: want{
				code: http.StatusBadRequest,
			},
			method:         http.MethodPost,
			request:        "/update/%s/%s/%s",
			metricTypeName: string(internal.CounterName),
			metricName:     "metric2",
			metricValue:    "value2",
		},
	}
	for _, tt := range gaugeUpdateMetricTests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			if strings.Count(tt.request, "%s") > 0 {
				target = fmt.Sprintf(tt.request, tt.metricTypeName, tt.metricName, tt.metricValue)
			} else {
				target = tt.request
			}
			request := httptest.NewRequest(tt.method, target, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(updateMetricHandler.ServeHTTP)
			h(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
	for _, tt := range counterMetricUpdateTests {
		t.Run(tt.name, func(t *testing.T) {
			target := fmt.Sprintf(tt.request, tt.metricTypeName, tt.metricName, tt.metricValue)
			request := httptest.NewRequest(tt.method, target, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(updateMetricHandler.ServeHTTP)
			h(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}
