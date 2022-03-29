package run

import (
	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
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
			response, _ := testRequest(t, ts, "POST", "/update/"+tt.url)
			if response.StatusCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, tt.want)
			}
			if tt.want == 200 {
				if handlers.Metrics.Get(tt.metricName) != tt.value {
					t.Errorf("Wrong metric value = %v, want %v", handlers.Metrics.Get(tt.metricName), tt.value)
				}
			}
			defer response.Body.Close()
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
				handlers.Metrics.Set("PollCount", strconv.FormatInt(tt.currentCountVal, 10))
			}
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			response, _ := testRequest(t, ts, "POST", "/update/"+tt.url)
			if response.StatusCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, tt.want)
			}
			defer response.Body.Close()
			if tt.want == 200 {
				if handlers.Metrics.Get(tt.metricName) != strconv.FormatInt(tt.count+tt.currentCountVal, 10) {
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

			response, _ := testRequest(t, ts, "POST", "/update/unknown/testCounter/100")
			defer response.Body.Close()

			if response.StatusCode != 501 {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, 501)
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		currMetrics map[string]string
		want        string
		statusCode  int
	}{
		{
			name:        "case 1. Counter metric",
			metricName:  "PollCount",
			currMetrics: map[string]string{"PollCount": "42"},
			want:        "42",
			statusCode:  200,
		},
		{
			name:        "case 2. Gauge metric",
			metricName:  "Alloc",
			currMetrics: map[string]string{"Alloc": "42.42"},
			want:        "42.42",
			statusCode:  200,
		},
		{
			name:        "case 3. Empty metric",
			metricName:  "HeadAlloc",
			currMetrics: map[string]string{"Alloc": "42.42"},
			statusCode:  404,
		},
		{
			name:        "case 4. Several metrics",
			metricName:  "BuckHashSys",
			currMetrics: map[string]string{"PollCount": "42", "Alloc": "42.42", "BuckHashSys": "123.0"},
			want:        "123.0",
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
					handlers.Metrics.Set(k, v)
				}
			}
			response, body := testRequest(t, ts, "GET", "/value/sometype/"+tt.metricName)
			defer response.Body.Close()
			if response.StatusCode != tt.statusCode {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, tt.statusCode)
			}
			if body != tt.want {
				t.Errorf("Wrong response = %v, want %v", body, tt.want)
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
			want:        "<html>\n\t\t\t<table>\n  \t\t\t\t<tr>\n    \t\t\t\t<td>Name</td>\n    \t\t\t\t<td>Value</td>\n  \t\t\t\t</tr>\n\t\t\t\t\n\t\t\t</table>\n\t\t</html>",
			statusCode:  200,
		},
		{
			name:        "case 2. One metric",
			currMetrics: map[string]string{"Alloc": "42.42"},
			want:        "<html>\n\t\t\t<table>\n  \t\t\t\t<tr>\n    \t\t\t\t<td>Name</td>\n    \t\t\t\t<td>Value</td>\n  \t\t\t\t</tr>\n\t\t\t\t\n\t\t<tr>\n\t\t\t<td>Alloc</td>\n\t\t\t<td>42.42</td>\n\t\t</tr>\n\t\t\t</table>\n\t\t</html>",
			statusCode:  200,
		},
		{
			name:        "case 3. Two metrics",
			currMetrics: map[string]string{"Alloc": "42.42", "PollCount": "42"},
			want:        "<html>\n\t\t\t<table>\n  \t\t\t\t<tr>\n    \t\t\t\t<td>Name</td>\n    \t\t\t\t<td>Value</td>\n  \t\t\t\t</tr>\n\t\t\t\t\n\t\t<tr>\n\t\t\t<td>Alloc</td>\n\t\t\t<td>42.42</td>\n\t\t</tr>\n\t\t<tr>\n\t\t\t<td>PollCount</td>\n\t\t\t<td>42</td>\n\t\t</tr>\n\t\t\t</table>\n\t\t</html>",
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
				handlers.Metrics.Set(k, v)
			}
			response, body := testRequest(t, ts, "GET", "/")
			defer response.Body.Close()
			if response.StatusCode != tt.statusCode {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, tt.statusCode)
			}
			if body != tt.want {
				t.Errorf("Wrong response = %v, want %v", body, tt.want)
			}
		})
	}
}
