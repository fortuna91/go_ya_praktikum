package run

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (int, []byte) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, respBody
}

func TestSetGaugeMetric(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		metricName string
		value      string
		want       int
	}{
		{
			name:       "case 1. Alloc 200 int",
			url:        "gauge/Alloc/123",
			metricName: "Alloc",
			value:      "123",
			want:       200,
		},
		{
			name:       "case 2. BuckHashSys 200 float",
			url:        "gauge/BuckHashSys/123.0",
			metricName: "BuckHashSys",
			value:      "123.0",
			want:       200,
		},
		{
			name:       "case 3. Other metric",
			url:        "gauge/MyMetric/123.456",
			metricName: "MyMetric",
			value:      "123.456",
			want:       200,
		},
		{
			name: "case 4. Wrong value",
			url:  "gauge/BuckHashSys/abc",
			want: 400,
		},
		{
			name: "case 5. Wrong value",
			url:  "gauge/BuckHashSys/",
			want: 404,
		},
		{
			name: "case 6. Wrong value",
			url:  "gauge/BuckHashSys",
			want: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			handlers.Metrics.ResetValues()
			responseCode, _ := testRequest(t, ts, "POST", "/update/"+tt.url, nil)
			if responseCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.want)
			}
			if tt.want == 200 {
				val, _ := strconv.ParseFloat(tt.value, 64)
				if *handlers.Metrics.Get(tt.metricName).Value != val {
					t.Errorf("Wrong metric value = %v, want %v", handlers.Metrics.Get(tt.metricName), tt.value)
				}
			}
		})
	}
}

func TestSetCountMetric(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		metricName      string
		count           int64
		currentCountVal int64
		want            int
	}{
		{
			name:            "case 1. Positive",
			url:             "counter/PollCount/1",
			metricName:      "PollCount",
			count:           1,
			currentCountVal: 0,
			want:            200,
		},
		{
			name:            "case 2. Other metric",
			url:             "counter/MyMetric/123",
			metricName:      "MyMetric",
			want:            200,
			count:           123,
			currentCountVal: 0,
		},
		{
			name: "case 3. Wrong value, empty",
			url:  "counter/PollCount/",
			want: 404,
		},
		{
			name: "case 4. Wrong value, str",
			url:  "counter/PollCount/abc",
			want: 400,
		},
		{
			name: "case 5. Wrong value, float",
			url:  "counter/PollCount/123.0",
			want: 400,
		},
		{
			name: "case 6. Wrong value, empty",
			url:  "counter/PollCount",
			want: 404,
		},
		{
			name:            "case 7. CurrCount",
			url:             "counter/PollCount/5",
			metricName:      "PollCount",
			count:           5,
			currentCountVal: 2,
			want:            200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// go Counter(CountChannel)
			if tt.want == 200 {
				handlers.Metrics.ResetValues()
				handlers.Metrics.UpdateCounter("PollCount", tt.currentCountVal)
			}
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			responseCode, _ := testRequest(t, ts, "POST", "/update/"+tt.url, nil)
			if responseCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.want)
			}
			if tt.want == 200 {
				if *handlers.Metrics.Get(tt.metricName).Delta != int64(tt.count+tt.currentCountVal) {
					t.Errorf("Wrong currCount = %v, want %v", handlers.Metrics.Get(tt.metricName), tt.count+tt.currentCountVal)
				}
			}
		})
	}
}

func TestNotImplemented(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want int
	}{
		{
			name: "case 1",
			url:  "/update/unknown/testCounter/100",
			want: 501,
		},
		{
			name: "case 2",
			url:  "/update/unknown/testCounter",
			want: 501,
		},
		{
			name: "case 3",
			url:  "/update/unknown",
			want: 501,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			responseCode, _ := testRequest(t, ts, "POST", "/update/unknown/testCounter/100", nil)

			if responseCode != 501 {
				t.Errorf("SendRequest() = %v, want %v", responseCode, 501)
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricType  string
		currMetrics map[string]string
		want        string
		statusCode  int
	}{
		{
			name:        "case 1. Counter metric",
			metricName:  "PollCount",
			metricType:  metrics.Counter,
			currMetrics: map[string]string{"PollCount": "42"},
			want:        "42",
			statusCode:  200,
		},
		{
			name:        "case 2. Gauge metric",
			metricName:  "Alloc",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"Alloc": "42.42"},
			want:        "42.420",
			statusCode:  200,
		},
		{
			name:        "case 3. Empty metric",
			metricName:  "HeadAlloc",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"Alloc": "42.42"},
			statusCode:  404,
		},
		{
			name:        "case 4. Several metrics",
			metricName:  "BuckHashSys",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"PollCount": "42", "Alloc": "42.42", "BuckHashSys": "123.01"},
			want:        "123.010",
			statusCode:  200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			if tt.statusCode == 200 {
				handlers.Metrics.ResetValues()
				for k, v := range tt.currMetrics {
					if tt.metricType == metrics.Gauge {
						val, _ := strconv.ParseFloat(v, 64)
						handlers.Metrics.SetGauge(k, &val)
					} else {
						val, _ := strconv.ParseInt(v, 10, 64)
						handlers.Metrics.UpdateCounter(k, val)
					}
				}
			}
			responseCode, body := testRequest(t, ts, "GET", "/value/sometype/"+tt.metricName, nil)
			if responseCode != tt.statusCode {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.statusCode)
			}
			if string(body) != tt.want {
				t.Errorf("Wrong response = %v, want %v", string(body), tt.want)
			}
		})
	}
}

func TestListMetrics(t *testing.T) {
	tests := []struct {
		name        string
		currMetrics map[string]string
		want        string
		statusCode  int
	}{
		{
			name:        "case 1. No metrics",
			currMetrics: map[string]string{},
			want:        "<!DOCTYPE html>\n\t<html>\n\t<table>\n\t<tr>\n\t<td>Name</td>\n\t<td>Value</td>\n\t</tr>\n\t</table>\n\t</html>",
			statusCode:  200,
		},
		{
			name:        "case 2. One metric",
			currMetrics: map[string]string{"Alloc": "42.42"},
			want:        "<!DOCTYPE html>\n\t<html>\n\t<table>\n\t<tr>\n\t<td>Name</td>\n\t<td>Value</td>\n\t</tr>\n\t<tr>\n\t<td>Alloc</td>\n\t<td>gauge</td>\n\t<td>&lt;nil&gt;</td>\n\t<td>42.42</td>\n\t</tr>\n\t</table>\n\t</html>",
			statusCode:  200,
		},
		{
			name:        "case 3. Two metrics",
			currMetrics: map[string]string{"Alloc": "42.42", "PollCount": "42"},
			want:        "<!DOCTYPE html>\n\t<html>\n\t<table>\n\t<tr>\n\t<td>Name</td>\n\t<td>Value</td>\n\t</tr>\n\t<tr>\n\t<td>Alloc</td>\n\t<td>gauge</td>\n\t<td>&lt;nil&gt;</td>\n\t<td>42.42</td>\n\t</tr>\n\t<tr>\n\t<td>PollCount</td>\n\t<td>gauge</td>\n\t<td>&lt;nil&gt;</td>\n\t<td>42</td>\n\t</tr>\n\t</table>\n\t</html>",
			statusCode:  200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			handlers.Metrics.ResetValues()
			for k, v := range tt.currMetrics {
				val, _ := strconv.ParseFloat(v, 64)
				handlers.Metrics.SetGauge(k, &val)
			}
			responseCode, body := testRequest(t, ts, "GET", "/", nil)
			if responseCode != tt.statusCode {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.statusCode)
			}
			if string(body) != tt.want {
				t.Errorf("Wrong response = %v, want %v", string(body), tt.want)
			}
		})
	}
}

func TestSetGaugeMetricJSON(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		want       int
	}{
		{
			name:       "case 1. Alloc 200 int",
			metricName: "Alloc",
			value:      123,
			want:       200,
		},
		{
			name:       "case 2. BuckHashSys 200 float",
			metricName: "BuckHashSys",
			value:      123.01,
			want:       200,
		},
		{
			name:       "case 3. Other metric",
			metricName: "MyMetric",
			value:      123.456,
			want:       200,
		},
		{
			name: "case 4. Wrong empty id",
			want: 400,
		},
		{
			name:       "case 5. Wrong empty value",
			metricName: "Alloc",
			want:       400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers.Metrics.ResetValues()

			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			val := &tt.value
			if tt.value == 0 {
				val = nil
			}
			metricReq := metrics.Metric{ID: tt.metricName, MType: metrics.Gauge, Value: val}
			body, _ := json.Marshal(metricReq)
			responseCode, _ := testRequest(t, ts, http.MethodPost, "/update", bytes.NewReader(body))
			if responseCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.want)
			}
			if tt.want == 200 {
				fmt.Println()
				fmt.Println(handlers.Metrics.Get(tt.metricName))
				if *handlers.Metrics.Get(tt.metricName).Value != tt.value {
					t.Errorf("Wrong metric value = %v, want %v", handlers.Metrics.Get(tt.metricName), tt.value)
				}
			}
		})
	}
}

func TestSetCountMetricJSON(t *testing.T) {
	tests := []struct {
		name            string
		metricName      string
		count           int64
		currentCountVal int64
		want            int
	}{
		{
			name:            "case 1. Positive",
			metricName:      "PollCount",
			count:           1,
			currentCountVal: 0,
			want:            200,
		},
		{
			name:            "case 2. Other metric",
			metricName:      "MyMetric",
			want:            200,
			count:           123,
			currentCountVal: 0,
		},
		{
			name:       "case 3. Wrong value, empty",
			metricName: "MyMetric",
			want:       400,
		},
		{
			name: "case 4. Wrong id, empty",
			want: 400,
		},
		{
			name:            "case 5. CurrCount",
			metricName:      "PollCount",
			count:           5,
			currentCountVal: 2,
			want:            200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == 200 {
				handlers.Metrics.ResetValues()
				handlers.Metrics.UpdateCounter("PollCount", tt.currentCountVal)
			}
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			count := &tt.count
			if tt.count == 0 {
				count = nil
			}
			metricReq := metrics.Metric{ID: tt.metricName, MType: metrics.Counter, Delta: count}
			body, _ := json.Marshal(metricReq)
			responseCode, _ := testRequest(t, ts, http.MethodPost, "/update", bytes.NewReader(body))
			if responseCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.want)
			}
			if tt.want == 200 {
				if *handlers.Metrics.Get(tt.metricName).Delta != int64(tt.count+tt.currentCountVal) {
					t.Errorf("Wrong currCount = %v, want %v", *handlers.Metrics.Get(tt.metricName).Delta, tt.count+tt.currentCountVal)
				}
			}
		})
	}
}

func TestGetMetricJSON(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricType  string
		currMetrics map[string]string
		wantDelta   int64
		wantValue   float64
		statusCode  int
	}{
		{
			name:        "case 1. Counter metric",
			metricName:  "PollCount",
			metricType:  metrics.Counter,
			currMetrics: map[string]string{"PollCount": "42"},
			wantDelta:   42,
			statusCode:  200,
		},
		{
			name:        "case 2. Gauge metric",
			metricName:  "Alloc",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"Alloc": "42.42"},
			wantValue:   42.42,
			statusCode:  200,
		},
		{
			name:        "case 3. Empty metric",
			metricName:  "HeadAlloc",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"Alloc": "42.42"},
			statusCode:  404,
		},
		{
			name:        "case 4. Several metrics",
			metricName:  "BuckHashSys",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"PollCount": "42", "Alloc": "42.42", "BuckHashSys": "123.01"},
			wantValue:   123.01,
			statusCode:  200,
		},
		{
			name:        "case 5. Empty metric name",
			metricType:  metrics.Gauge,
			currMetrics: map[string]string{"Alloc": "42.42"},
			statusCode:  400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			if tt.statusCode == 200 {
				handlers.Metrics.ResetValues()
				for k, v := range tt.currMetrics {
					if tt.metricType == metrics.Gauge {
						val, _ := strconv.ParseFloat(v, 64)
						handlers.Metrics.SetGauge(k, &val)
					} else {
						val, _ := strconv.ParseInt(v, 10, 64)
						handlers.Metrics.UpdateCounter(k, val)
					}
				}
			}
			metricReq := metrics.Metric{ID: tt.metricName, MType: metrics.Gauge}
			body, _ := json.Marshal(metricReq)
			responseCode, bodyResp := testRequest(t, ts, http.MethodPost, "/value", bytes.NewReader(body))
			if responseCode != tt.statusCode {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.statusCode)
			}
			metricResp := metrics.Metric{}
			json.Unmarshal(bodyResp, &metricResp)

			if tt.statusCode == 200 {
				if tt.metricType == metrics.Gauge {
					if *metricResp.Value != tt.wantValue {
						t.Errorf("Wrong response = %v, want %v", body, tt.wantValue)
					}
				} else if tt.metricType == metrics.Counter {
					if *metricResp.Delta != tt.wantDelta {
						t.Errorf("Wrong response = %v, want %v", body, tt.wantDelta)
					}
				}
			}
		})
	}
}
