package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestStorageStateHandler_ServeHTTP(t *testing.T) {
	var gaugeStorage storage.MemStorage[internal.Gauge]
	var counterStorage storage.MemStorage[internal.Counter]
	gaugeStorage.Init()
	counterStorage.Init()

	storageStateHandler := StorageStateHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	srv := httptest.NewServer(&storageStateHandler)
	defer srv.Close()

	type want struct {
		code         int
		responseBody string
		contentType  string
	}
	type stateTestCase struct {
		name   string
		method string
		want   want
	}
	stateTests := []stateTestCase{
		{
			name:   "success",
			method: http.MethodGet,
			want: want{
				code:         http.StatusOK,
				responseBody: "\n",
				contentType:  "text/plain",
			},
		},
		{
			name:   "gauge wrong method post",
			method: http.MethodPost,
			want:   want{code: http.StatusMethodNotAllowed},
		},
		{
			name:   "gauge wrong method put",
			method: http.MethodPut,
			want:   want{code: http.StatusMethodNotAllowed},
		},
		{
			name:   "gauge wrong method delete",
			method: http.MethodDelete,
			want:   want{code: http.StatusMethodNotAllowed},
		},
	}
	for _, tt := range stateTests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.want.code, resp.StatusCode(), "Response code didn't match expected")
			if tt.want.responseBody != "" {
				assert.Equal(t, tt.want.responseBody, string(resp.Body()))
			}
		})
	}
}

func TestUpdateMetricHandler_ServeHTTP(t *testing.T) {
	var gaugeStorage storage.MemStorage[internal.Gauge]
	var counterStorage storage.MemStorage[internal.Counter]
	gaugeStorage.Init()
	counterStorage.Init()

	type want struct {
		code int
	}
	type updateMetricTestCase struct {
		name           string
		want           want
		method         string
		request        string
		metricTypeName string
		metricName     string
		metricValue    string
	}

	updateMetricHandler := UpdateMetricHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	r := chi.NewRouter()
	r.Handle("/update/{metricType}/{metricName}/{metricValue}", &updateMetricHandler)
	updateMetricSrv := httptest.NewServer(r)
	defer updateMetricSrv.Close()

	updateMetricTests := []updateMetricTestCase{
		{
			name: "success gauge update",
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
			want: want{
				code: http.StatusNotFound,
			},
			method:  http.MethodPost,
			request: fmt.Sprintf("/update/%s/%s", string(internal.GaugeName), strconv.FormatFloat(123.45, 'f', -1, 64)),
		},
		{
			name: "gauge wrong value",
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
			want: want{
				code: http.StatusBadRequest,
			},
			method:         http.MethodPost,
			request:        "/update/%s/%s/%s",
			metricTypeName: "unexpected",
			metricName:     "metric1",
			metricValue:    strconv.FormatFloat(123.45, 'f', -1, 64),
		},
		{
			name: "counter success update",
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
	for _, tt := range updateMetricTests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			if strings.Count(tt.request, "%s") > 0 {
				target = fmt.Sprintf(tt.request, tt.metricTypeName, tt.metricName, tt.metricValue)
			} else {
				target = tt.request
			}
			req := resty.New().R()
			req.Method = tt.method
			req.URL = fmt.Sprintf("%s%s", updateMetricSrv.URL, target)

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.want.code, resp.StatusCode(), "Response code didn't match expected")
		})
	}
}

func TestMetricStateHandler_ServeHTTP(t *testing.T) {
	var gaugeStorage storage.MemStorage[internal.Gauge]
	var counterStorage storage.MemStorage[internal.Counter]
	gaugeStorage.Init()
	counterStorage.Init()

	metricName1 := "metric1"
	metricValue1 := internal.Gauge(123)
	metricName2 := "metric2"
	metricValue2 := internal.Counter(123)
	metricName3 := "metric3"
	gaugeStorage.Set(metricName1, metricValue1)
	counterStorage.Set(metricName2, metricValue2)

	metricStateHandler := MetricStateHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	r := chi.NewRouter()
	r.Handle("/value/{metricType}/{metricName}", &metricStateHandler)
	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		code         int
		responseBody string
		contentType  string
	}
	type metricStateTestCase struct {
		name       string
		method     string
		metricName string
		metricType string
		want       want
	}
	metricStateTests := []metricStateTestCase{
		{
			name:       "success gauge",
			method:     http.MethodGet,
			metricName: metricName1,
			metricType: string(internal.GaugeName),
			want: want{
				code:         http.StatusOK,
				responseBody: metricValue1.String(),
				contentType:  "text/plain",
			},
		},
		{
			name:       "success counter",
			method:     http.MethodGet,
			metricName: metricName2,
			metricType: string(internal.CounterName),
			want: want{
				code:         http.StatusOK,
				responseBody: metricValue2.String(),
				contentType:  "text/plain",
			},
		},
		{
			name:       "metric not found",
			method:     http.MethodGet,
			metricName: metricName3,
			metricType: string(internal.GaugeName),
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:       "wrong metric type",
			method:     http.MethodGet,
			metricName: metricName1,
			metricType: "unexpected",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			metricName: metricName1,
			metricType: string(internal.GaugeName),
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range metricStateTests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = fmt.Sprintf("%s/value/%s/%s", srv.URL, tt.metricType, tt.metricName)

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.want.code, resp.StatusCode(), "Response code didn't match expected")
			if tt.want.responseBody != "" {
				assert.Equal(t, tt.want.responseBody, string(resp.Body()))
			}
		})
	}
}
