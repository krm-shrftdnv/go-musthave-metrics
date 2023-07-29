package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/serializer"
	"github.com/krm-shrftdnv/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
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
				responseBody: "[]",
				contentType:  "application/json",
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
	metricValue1 := internal.Gauge(123.45)
	metricValue2 := internal.Counter(123)
	metric1 := serializer.Metrics{
		ID:    "metric1",
		MType: string(internal.GaugeName),
		Value: &metricValue1,
	}
	metric2 := serializer.Metrics{
		ID:    "metric2",
		MType: string(internal.CounterName),
		Delta: &metricValue2,
	}

	type want struct {
		code         int
		responseBody *serializer.Metrics
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
	r.Handle("/update", &updateMetricHandler)
	updateMetricSrv := httptest.NewServer(r)
	defer updateMetricSrv.Close()

	updateMetricTests := []updateMetricTestCase{
		{
			name: "success gauge update",
			want: want{
				code:         http.StatusOK,
				responseBody: &metric1,
			},
			method:         http.MethodPost,
			request:        "/update",
			metricTypeName: string(internal.GaugeName),
			metricName:     metric1.ID,
			metricValue:    strconv.FormatFloat(float64(*metric1.Value), 'f', -1, 64),
		},
		{
			name: "gauge wrong method",
			want: want{
				code: http.StatusMethodNotAllowed,
			},
			method:         http.MethodGet,
			request:        "/update",
			metricTypeName: string(internal.GaugeName),
			metricName:     metric1.ID,
			metricValue:    strconv.FormatFloat(float64(*metric1.Value), 'f', -1, 64),
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
			request:        "/update",
			metricTypeName: string(internal.GaugeName),
			metricName:     metric1.ID,
			metricValue:    "value1",
		},
		{
			name: "gauge wrong type",
			want: want{
				code: http.StatusBadRequest,
			},
			method:         http.MethodPost,
			request:        "/update",
			metricTypeName: "unexpected",
			metricName:     metric1.ID,
			metricValue:    strconv.FormatFloat(float64(*metric1.Value), 'f', -1, 64),
		},
		{
			name: "counter success update",
			want: want{
				code:         http.StatusOK,
				responseBody: &metric2,
			},
			method:         http.MethodPost,
			request:        "/update",
			metricTypeName: string(internal.CounterName),
			metricName:     metric2.ID,
			metricValue:    strconv.FormatInt(int64(*metric2.Delta), 10),
		},
		{
			name: "counter wrong value",
			want: want{
				code: http.StatusBadRequest,
			},
			method:         http.MethodPost,
			request:        "/update",
			metricTypeName: string(internal.CounterName),
			metricName:     metric2.ID,
			metricValue:    "value2",
		},
	}
	for _, tt := range updateMetricTests {
		t.Run(tt.name, func(t *testing.T) {
			var metric serializer.Metrics
			req := resty.New().R()

			value, err := strconv.ParseFloat(tt.metricValue, 64)
			if err != nil {
				metricStringValue := struct {
					serializer.Metrics
					Value string `json:"value"`
				}{
					Metrics: serializer.Metrics{
						ID:    tt.metricName,
						MType: tt.metricTypeName,
					},
					Value: tt.metricValue,
				}
				req.SetBody(metricStringValue)
			} else {
				switch tt.metricTypeName {
				case string(internal.GaugeName):
					gaugeValue := internal.Gauge(value)
					metric = serializer.Metrics{
						ID:    tt.metricName,
						MType: tt.metricTypeName,
						Value: &gaugeValue,
					}
				case string(internal.CounterName):
					counterValue := internal.Counter(value)
					metric = serializer.Metrics{
						ID:    tt.metricName,
						MType: tt.metricTypeName,
						Delta: &counterValue,
					}
				}
				req.SetBody(metric)
			}
			req.Method = tt.method
			req.SetHeader("Content-Type", "application/json")
			req.URL = fmt.Sprintf("%s%s", updateMetricSrv.URL, tt.request)

			resp, err := req.Send()
			assert.Equal(t, tt.want.code, resp.StatusCode(), "Response code didn't match expected")
			if tt.want.code == http.StatusOK {
				assert.NoError(t, err, "error making HTTP request")
				var metric serializer.Metrics
				assert.NoError(t, err, "error reading response body")
				err = json.Unmarshal(resp.Body(), &metric)
				assert.NoError(t, err, "error unmarshalling response body")
				ok := reflect.DeepEqual(tt.want.responseBody, &metric)
				assert.True(t, ok, "response body didn't match expected")
			}
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
	metric1 := serializer.Metrics{
		ID:    metricName1,
		MType: string(internal.GaugeName),
		Value: &metricValue1,
	}
	metric2 := serializer.Metrics{
		ID:    metricName2,
		MType: string(internal.CounterName),
		Delta: &metricValue2,
	}

	metricStateHandler := MetricStateHandler{
		GaugeStorage:   &gaugeStorage,
		CounterStorage: &counterStorage,
	}
	r := chi.NewRouter()
	r.Handle("/value", &metricStateHandler)
	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		code         int
		responseBody *serializer.Metrics
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
			method:     http.MethodPost,
			metricName: metricName1,
			metricType: string(internal.GaugeName),
			want: want{
				code:         http.StatusOK,
				responseBody: &metric1,
				contentType:  "text/plain",
			},
		},
		{
			name:       "success counter",
			method:     http.MethodPost,
			metricName: metricName2,
			metricType: string(internal.CounterName),
			want: want{
				code:         http.StatusOK,
				responseBody: &metric2,
				contentType:  "text/plain",
			},
		},
		{
			name:       "metric not found",
			method:     http.MethodPost,
			metricName: metricName3,
			metricType: string(internal.GaugeName),
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:       "wrong metric type",
			method:     http.MethodPost,
			metricName: metricName1,
			metricType: "unexpected",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
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
			req.URL = fmt.Sprintf("%s/value", srv.URL)
			req.SetBody(serializer.Metrics{
				ID:    tt.metricName,
				MType: tt.metricType,
			})

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.want.code, resp.StatusCode(), "Response code didn't match expected")
			if tt.want.responseBody != nil {
				var metric serializer.Metrics
				assert.NoError(t, err, "error reading response body")
				err = json.Unmarshal(resp.Body(), &metric)
				assert.NoError(t, err, "error unmarshalling response body")
				ok := reflect.DeepEqual(tt.want.responseBody, &metric)
				assert.True(t, ok, "response body didn't match expected")
			}
		})
	}
}
